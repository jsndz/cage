A **veth (virtual ethernet)** device is a software-only network interface.

It behaves like a NIC:

```bash
ip addr
ip link
tcpdump
ping
```

all work on it.

But there is no hardware behind it.

---

For a physical NIC:

PCI NIC Card
    |
    +-- Ethernet Port

the kernel creates a net_device such as:

eth0
ens33
enp1s0

That net_device represents the NIC's networking interface. When packets are sent through eth0, the NIC driver eventually programs the hardware to transmit them on the physical Ethernet port.

---

A **veth pair** is two virtual interfaces connected together.

```text id="63jj50"
vethA <------> vethB
```

Think of it as a virtual Ethernet cable.

If a packet enters one side:

```text id="zkm95k"
vethA
  ↓
kernel
  ↓
vethB
```

it immediately appears on the other side.

---

Creating one:

```bash
ip link add vethA type veth peer name vethB
```

Kernel creates:

```text id="mjgr4e"
vethA
vethB
```

as a connected pair.

---

At the kernel level, both are:

```c
struct net_device
```

objects.

Conceptually:

```c
struct veth {
    struct net_device *peer;
};
```

When transmitting:

```c
vethA->send(packet)
```

the driver does roughly:

```c
deliver(packet, vethB);
```

instead of sending to hardware.

---

Physical NIC:

```text id="0kwj7j"
Application
    ↓
Socket
    ↓
Kernel
    ↓
NIC Driver
    ↓
DMA
    ↓
Network Card
    ↓
Wire
```

Veth:

```text id="gzyx8t"
Application
    ↓
Socket
    ↓
Kernel
    ↓
veth Driver
    ↓
Peer Interface
```

No wire.

No hardware.

No DMA.

Just memory operations inside the kernel.

---

You can put each side in different namespaces.

Example:

```text id="vnjsps"
Host Namespace
    veth-host

Container Namespace
    veth-container
```

Commands:

```bash
ip netns add ns1

ip link add veth-host type veth peer name veth-container

ip link set veth-container netns ns1
```

Result:

```text id="1zknjd"
Host
  veth-host
      |
      |
      |
Container
  veth-container
```

Now they can communicate directly.

---

Assign IPs:

Host:

```bash
ip addr add 10.0.0.1/24 dev veth-host
ip link set veth-host up
```

Container:

```bash
ip netns exec ns1 ip addr add 10.0.0.2/24 dev veth-container
ip netns exec ns1 ip link set veth-container up
```

Now:

```bash
ping 10.0.0.2
```

works.

---

This is how Docker containers are connected.

```text id="3a6d68"
Container
   eth0
     |
veth-container
     |
veth-host
     |
docker0 bridge
     |
Host NIC
     |
Internet
```

Every container gets its own veth pair.

---

Important property:

If one side goes down:

```bash
ip link set vethA down
```

the peer sees:

```text id="dzkffn"
NO-CARRIER
```

just like unplugging a real cable.

Because the kernel models it as a physical Ethernet link.

---

A packet moving through a veth pair is still a full Ethernet frame:

```text id="1h3b4d"
+------------------+
| Ethernet Header  |
+------------------+
| IP Header        |
+------------------+
| TCP Header       |
+------------------+
| Data             |
+------------------+
```

The veth driver doesn't strip or reinterpret it.

It simply hands the frame to the peer interface.

---

The mental model is:

```text id="p9m0t6"
Physical cable:
NIC <------wire------> NIC

Virtual cable:
vethA <----kernel----> vethB
```

A veth pair exists primarily to connect different network namespaces while making each side look like a normal Ethernet interface.




`eth` comes from **Ethernet**, but Ethernet is more than just a cable.

Ethernet defines:

* Frame format
* MAC addresses
* Layer 2 communication rules

Originally it was used over physical cables, which is why Linux named interfaces:

```text
eth0
eth1
eth2
```

---

Wi-Fi is actually different.

A Wi-Fi card typically appears as:

```text
wlan0
```

or on modern systems:

```text
wlp2s0
```

where:

```text
wlan = Wireless LAN
```

Example:

```text
Wi-Fi Card
   ↓
wlan0
```

---

Why are containers usually given `eth0`?

Because a veth pair behaves like an Ethernet link.

The packets sent through a veth are Ethernet frames:

```text
Container eth0
     ↓
Ethernet Frame
     ↓
veth
     ↓
Host
```

So naming it `eth0` makes sense.

---

Suppose your laptop uses Wi-Fi:

```text
Internet
    ↓
Wi-Fi Router
    ↓
Laptop Wi-Fi Card (wlan0)
```

Docker may create:

```text
Container
   eth0
     ↓
veth
     ↓
Bridge
     ↓
Host Networking
     ↓
wlan0
     ↓
Wi-Fi
```

The container never talks directly to the Wi-Fi hardware.

It only sees a virtual Ethernet interface (`eth0`).

The host kernel translates and forwards packets to the real Wi-Fi device.

---

At the device level:

```text
Container
    eth0 (virtual Ethernet)
        ↓
veth
        ↓
Host Network Stack
        ↓
wlan0 (real Wi-Fi device)
        ↓
802.11 Wi-Fi frames
        ↓
Router
```

Notice that:

```text
Container: Ethernet frames
Host Wi-Fi card: Wi-Fi frames
```

The Linux networking stack converts between them.

---

A useful analogy:

```text
Application
    ↓
IP Packet
```

The application doesn't care whether the packet leaves via:

```text
eth0   (Ethernet cable)
wlan0  (Wi-Fi)
tun0   (VPN)
veth0  (virtual cable)
```

The network device driver handles the details.

That's why containers almost always get an `eth0` even when the host machine is connected to the Internet through Wi-Fi. The container sees a virtual Ethernet network, and the host takes care of getting those packets onto the actual Wi-Fi network.
