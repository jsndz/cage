Without Namespace you have:

NIC
 ↓
Kernel
 ↓
Network namespace
 ↓
Routing table
 ↓
Socket

The kernel first determines which network namespace owns the interface that received the packet.

Then all networking operations happen inside that namespace's struct net.

But with namespace you have:


struct net #1  (host)
struct net #2  (container A)
struct net #3  (container B)

Network namespace gives you sepearate network kernel struct. ie, spearate network stack
```md 
struct net {
    routing_tables
    interfaces
    conntrack
    firewall_rules
    ...
};
```
The physical NIC is attached to the host namespace

Host Namespace
 └── Physical NIC
      (eth0)


if network stack of container and host is completely isolated 
So when the data comes to the NIC it goes to the host namespace.

So data comes for namespace B 
so how will data be received to container,
using, ethernet pairs

ethernet pairs are used to connect namespaces

ip link add veth0 type veth peer name veth1

Kernel creates:

veth0 <----> veth1

should move one end to the container

The kernel treats this pair like a cable.

Anything sent into one end appears at the other end.

This is very similar to:

PC Ethernet Port  <=======>  Router Ethernet Port

except both ends exist inside the same kernel.

Host
 ├─ eth0 (physical NIC)
 ├─ docker0 bridge
 └─ veth-host
        ||
        || virtual cable
        ||
 Container
 └─ eth0


 A network interface like:

eth0
lo
veth123
exists as a struct net_device which is exposed through /sys/class/net/

veth device have peers and if any data comes they emit to the peer recv path

When you run:

ip link add veth0 type veth peer name veth1

the kernel:

Allocates two net_device structures.
Links them as peers.
Registers them with the networking subsystem.
Places them into namespaces.
Exposes them through ip link.

A network namespace (struct net) does not have an IP address.

net_device is the kernel's representation of a network interface. Which has IP Address
The IP address belongs to the network interface (net_device), not the physical NIC hardware.
eth0
wlan0
lo
docker0
veth123

Each of these corresponds to a struct net_device in kernel memory.

The kernel's net_device stores:

IP addresses
MTU
routes association
statistics
driver callbacks

so you can have multiple network devices
Physical NIC
    |
    +-- eth0  (192.168.1.10)
    +-- eth0.100 (VLAN)
    +-- eth0.200 (VLAN)

struct net = Network Namespace
struct net_device = Network Interface
Represents one interface inside a namespace.

You do not move wlan0 into the container.

Instead:

Container
   eth0 (veth)
      ↓
Host
   wlan0
      ↓
Wi-Fi Router

The container sees only its virtual eth0, and the host forwards traffic through wlan0.

This is how almost all containers get Internet access.