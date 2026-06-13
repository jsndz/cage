A Linux bridge is a **software switch** implemented inside the kernel.

Think of it as the virtual equivalent of a network switch.

Physical switch:

```text id="7vh0i6"
PC1 ----+
        |
PC2 ----+---- Switch
        |
PC3 ----+
```

Linux bridge:

```text id="86j09u"
veth1 ---+
         |
veth2 ---+--- br0
         |
veth3 ---+
```

`br0` is the bridge.

---

Without a bridge:

```text id="pryiyh"
Container A
   |
vethA

Container B
   |
vethB
```

These interfaces are independent.

Packets sent from A have nowhere to go unless you manually route them.



Just with veth pairs it can be a one to one connection like host to container
If every container needs to be interconnected then you need mesh connection which is less than ideal

so use bridge think of it like a bus every networks connects(bus topology)
---

With a bridge:

```text id="fylhcb"
Container A
   |
vethA
   |
vethA-host
      \
       \
        br0
       /
      /
vethB-host
   |
vethB
   |
Container B
```

Now the bridge forwards frames between them.

---

How does a bridge work?

Every Ethernet device has a MAC address:

```text id="hfwmpr"
Container A: AA:AA:AA
Container B: BB:BB:BB
```

When a frame arrives:

```text id="20gcot"
src = AA:AA:AA
dst = BB:BB:BB
```

the bridge learns:

```text id="yxjlwm"
AA:AA:AA -> port 1
```

and stores it in a forwarding database (FDB).

Conceptually:

```text id="3iw2qn"
MAC Address     Port
-----------     ----
AA:AA:AA        vethA
BB:BB:BB        vethB
```

---

When a frame comes in:

```text id="r0j0dj"
Destination = BB:BB:BB
```

the bridge checks:

```text id="ix2spz"
FDB lookup
```

and forwards only to the correct port.

Exactly like a physical switch.

---

If the destination MAC is unknown:

```text id="clo4oc"
Destination = CC:CC:CC
```

the bridge floods:

```text id="sj4sni"
send to all ports except incoming port
```

until it learns where CC lives.

Again, same as a real switch.

---

Why do containers need bridges?

Suppose you have:

```text id="gvafji"
Container A
Container B
Container C
```

You could create direct veth pairs:

```text id="p0r1kh"
A <-> B
A <-> C
B <-> C
```

but that becomes:

```text id="6q7b5n"
n*(n-1)/2 links
```

which doesn't scale.

Instead:

```text id="snn55x"
A
 \
  \
   br0
  /
 /
B

 \
  C
```

Every container connects to one bridge.

---

Docker's default setup:

```text id="2nf1z6"
Container A
    |
 vethA
    |
 vethA-host
    |
 docker0
    |
 Host
```

`docker0` is just a Linux bridge.

All container host-side veth interfaces attach to it.

---

Bridge + Internet access:

```text id="gc6hww"
Container
   |
veth
   |
bridge
   |
Host NIC
   |
Internet
```

The bridge connects container interfaces into the same Layer 2 network.

Then the host can perform:

* Routing
* NAT
* Firewalling

to reach the Internet.

---

A bridge operates at Layer 2.

It looks only at:

```text id="8n6fdu"
MAC addresses
Ethernet frames
```

It does **not** care about:

```text id="8c8kv9"
IP addresses
TCP ports
```

Those are Layer 3 and Layer 4 concepts handled later.

---

In your `cage` runtime, a common setup would be:

```text id="w3jjlwm"
Container
   |
eth0
   |
veth-container
   |
veth-host
   |
cage0 (bridge)
   |
Host NIC
```

The bridge acts like a virtual switch connecting all container interfaces together and providing a common point from which packets can be routed or NATed to the outside world.
