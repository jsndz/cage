now container has the ip address with eth0 net_device
but it is private address
and not recognisable since it is private
you need to put it to NAT so that it can be marked with public ip
nftables is one of the tools that implements NAT rules in the Linux kernel.

A table is just a logical container for chains and rules.

table
   └── chains
          └── rules


Here you will observe chain  hooks: like forward, postrouting and accept

These are hooks in the Linux networking stack.

When a packet moves through the kernel, it passes through specific stages.

Simplified flow:

Packet arrives
      ↓
PREROUTING
      ↓
Routing Decision
      ↓
INPUT      (for local machine)
or
FORWARD    (for another machine)
      ↓
POSTROUTING
      ↓
Packet leaves
FORWARD

The FORWARD hook is used when the packet is passing through the host.

Example:

Container
 10.0.0.2
    ↓
 Host
    ↓
 Google

The packet is not for the host.

The host is acting like a router.

So the packet enters the:

FORWARD

hook.

Your rule:

iif=cage0
oif=eth0
accept

means:

If packet came from cage0
and is leaving through eth0

allow it.

Without this:

Container -> Internet

gets dropped.

POSTROUTING

POSTROUTING happens after Linux has decided where the packet will go.

Example:

10.0.0.2 -> 8.8.8.8

Linux routing table determines:

Send through eth0

Only after that does the packet enter:

POSTROUTING

At this point you know:

Outgoing interface = eth0

which is why NAT is usually done here.

Your rule:

ip saddr 10.0.0.0/24
oifname eth0
masquerade

runs in POSTROUTING.

Before:

SRC=10.0.0.2
DST=8.8.8.8

After:

SRC=HostIP
DST=8.8.8.8

Then packet leaves.