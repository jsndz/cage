
### Without namespaces

You have:

```text
Physical NIC
   ↓
Kernel networking stack
   ↓
Routing table
   ↓
Socket
```

When a packet arrives:

1. NIC receives Ethernet frame.
2. NIC DMA-copies packet into RAM.
3. NIC raises interrupt (or NAPI polling picks it up).
4. Kernel networking stack processes packet.
5. Kernel finds matching socket.
6. Process receives data.

Only one networking stack exists.

---

### With network namespaces

Linux does **not** create another physical networking stack in hardware.

There is still:

```text
One NIC
One CPU
One kernel
```

The kernel creates multiple logical networking contexts:

```text
struct net #1  (host)
struct net #2  (container A)
struct net #3  (container B)
```

Each context has:

```text
own routes
own interfaces
own ports
own firewall rules
```

---

### Where does the NIC belong?

Normally:

```text
Host Namespace
 └── Physical NIC
      (eth0)
```

The physical NIC is attached to the host namespace.

When a packet arrives:

```text
Internet
   ↓
Physical NIC
   ↓
Host namespace
```

The packet always enters through the namespace that owns the NIC.

---

### How does the container receive packets then?

Using a virtual ethernet pair:

```text
Host namespace                    Container namespace

veth-host  <------->  veth-container
```

The kernel treats this pair like a cable.

---

Suppose container sends a packet:

```text
Container Process
        ↓
Socket
        ↓
veth-container
```

The veth driver receives it.

Kernel immediately injects the packet into:

```text
veth-host
```

on the other side.

Conceptually:

```text
write(veth-container)
          ↓
Kernel memory copy
          ↓
read(veth-host)
```

No actual wire exists.

---

### Then where does the physical NIC come in?

Usually there is a bridge:

```text
              bridge
             /      \
        veth-host   eth0
```

Like a virtual switch.

Packet flow:

```text
Container
   ↓
veth-container
   ↓
veth-host
   ↓
Bridge
   ↓
Physical NIC
   ↓
Internet
```

---

### What happens on receive?

Packet arrives from Internet:

```text
Internet
   ↓
Physical NIC
   ↓
Bridge
   ↓
veth-host
   ↓
veth-container
   ↓
Container socket
```

The kernel forwards it between virtual devices exactly as a switch would.

---

### Device-level view

Physical NIC:

```text
PCI Device
 ├── DMA engine
 ├── RX queue
 ├── TX queue
 └── MAC address
```

Kernel object:

```c
struct net_device
```

represents it.

---

Virtual interface (`veth`):

```text
No hardware
No PCI device
No DMA
```

Just a kernel object:

```c
struct net_device
```

plus code that says:

```c
Transmit(packet) {
    deliver_to_peer(packet);
}
```

So a veth behaves like a NIC from userspace's perspective:

```bash
ip addr
ip route
tcpdump
```

all work.

But internally it's just software.

---

### Why can two namespaces both have eth0?

Because interface names are namespace-local.

Example:

```text
Host Namespace
    eth0

Container A
    eth0

Container B
    eth0
```

These are three different `struct net_device` objects.

The kernel lookup is effectively:

```text
(namespace, interface_name)
```

not just:

```text
interface_name
```

So:

```text
Host eth0         -> device #1
Container A eth0 -> device #2
Container B eth0 -> device #3
```

even though all are called `eth0`.

---

The key idea is that network namespaces do not virtualize the hardware. They virtualize the kernel's view of networking resources. The actual packet movement still happens through real or virtual network devices (`struct net_device`) connected together by kernel networking components such as veth pairs, bridges, routing, and NAT.
