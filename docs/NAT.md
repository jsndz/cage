NAT (Network Address Translation) rewrites packet addresses as they pass through a router.

It allows many private IPs to share one public IP.

---

Without NAT

Suppose:

```text id="76ah29"
Container: 10.0.0.2
Server:    8.8.8.8
```

Packet:

```text id="d9nhm6"
SRC=10.0.0.2
DST=8.8.8.8
```

reaches the Internet.

Problem:

```text id="shjlwm"
10.0.0.2
```

is a private IP. Internet routers cannot route replies back to it.

The packet gets dropped.

---

With NAT

Container sends:

```text id="j4z1ix"
SRC=10.0.0.2:50000
DST=8.8.8.8:443
```

At the host:

```text id="lvr06r"
Public IP = 203.0.113.5
```

NAT rewrites:

```text id="a9znrb"
SRC=203.0.113.5:60000
DST=8.8.8.8:443
```

and stores:

```text id="f2zof4"
203.0.113.5:60000
    ↔
10.0.0.2:50000
```

in a connection-tracking table.

---

Reply arrives:

```text id="2u5k6i"
SRC=8.8.8.8:443
DST=203.0.113.5:60000
```

Host looks up:

```text id="3f1z6w"
203.0.113.5:60000
```

in the NAT table and rewrites:

```text id="4dzw6n"
SRC=8.8.8.8:443
DST=10.0.0.2:50000
```

Then forwards it to the container.

Container thinks it talked directly to 8.8.8.8.

---

Conceptually:

```text id="s7vkg4"
Container
10.0.0.2
     |
     |
Host NAT
203.0.113.5
     |
Internet
```

Many containers can share one public IP.

---

Example:

```text id="fxr1pf"
Container A 10.0.0.2:50000
Container B 10.0.0.3:50000
```

Both connect to Google.

NAT creates:

```text id="v8ls4j"
10.0.0.2:50000 -> 203.0.113.5:60001
10.0.0.3:50000 -> 203.0.113.5:60002
```

Outside world sees:

```text id="uw3h77"
203.0.113.5
```

only.

---

Why ports are changed

If two containers use:

```text id="izt4pt"
10.0.0.2:50000
10.0.0.3:50000
```

and both became:

```text id="yqzjlwm"
203.0.113.5:50000
```

the host couldn't tell which reply belongs to which container.

So NAT often rewrites both:

* IP
* Source port

This is technically PAT (Port Address Translation), though people usually just call it NAT.

---

Docker networking uses NAT.

```text id="0jnqjp"
Container
10.0.0.2
    |
Bridge
    |
Host
192.168.1.10
    |
Wi-Fi Router
    |
Internet
```

Container sends:

```text id="nqzddt"
10.0.0.2 -> 8.8.8.8
```

Host NAT changes it:

```text id="wx1gto"
192.168.1.10 -> 8.8.8.8
```

Then your home router may NAT again:

```text id="u2nkx0"
49.x.x.x -> 8.8.8.8
```

This is called double NAT.

---

In Linux, NAT is usually implemented through:

* netfilter
* iptables
* nftables
* conntrack

The kernel maintains a connection-tracking table that remembers every active translation.

For your `cage` runtime, NAT is what allows containers with private addresses like:

```text id="7apukm"
10.0.0.2
10.0.0.3
10.0.0.4
```

to access the Internet through the host's single network connection without needing public IP addresses.
