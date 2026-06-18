Port forwarding means:

```text
Internet
    ↓
Host:8080
    ↓
Container:80
```

A service is running inside the container:

```text
10.0.0.2:80
```

but nobody outside can reach it because:

```text
10.0.0.2
```

is a private container IP.

So we expose it through the host:

```text
HostIP:8080
```

---

## Without port forwarding

Container:

```text
10.0.0.2:80
```

User:

```text
curl HostIP:8080
```

Kernel receives:

```text
DST=HostIP:8080
```

No process is listening on:

```text
HostIP:8080
```

Connection fails.

---

## With port forwarding

We tell Linux:

```text
Whenever a packet comes to Host:8080

change destination to

10.0.0.2:80
```

So:

```text
HostIP:8080
```

becomes:

```text
10.0.0.2:80
```

This destination rewrite is called:

```text
DNAT
Destination NAT
```

---

## PREROUTING

External traffic enters the host:

```text
Internet
    ↓
Host
```

Before Linux decides where to send the packet, it reaches:

```text
PREROUTING
```

There we do:

```text
8080
 ↓
10.0.0.2:80
```

Packet:

```text
HostIP:8080
```

becomes:

```text
10.0.0.2:80
```

Then Linux routes it to the container.

---

## OUTPUT

Suppose the request comes from the host itself:

```bash
curl localhost:8080
```

This packet never enters from the network.

It starts inside the host.

Flow:

```text
Host Process
      ↓
OUTPUT
      ↓
Routing
```

Since it never hits PREROUTING, we need another DNAT rule in:

```text
OUTPUT
```

so:

```text
localhost:8080
```

also becomes:

```text
10.0.0.2:80
```

---

## FORWARD

After DNAT:

```text
10.0.0.2:80
```

Linux decides:

```text
Send packet to cage0 bridge
```

Packet now passes through:

```text
FORWARD
```

hook.

If firewall policy is:

```text
DROP
```

we must explicitly allow:

```text
destination 10.0.0.2
port 80
```

using a FORWARD rule.

---

## Why you don't currently need Rule 4

Your current nftables code uses:

```go
Policy: ACCEPT
```

for the FORWARD chain.

That means:

```text
Everything is already allowed
```

So:

```text
FORWARD allow 10.0.0.2:80
```

is redundant right now.

---

## Full packet journey

User:

```text
curl HostIP:8080
```

Packet:

```text
DST=HostIP:8080
```

### Step 1

PREROUTING:

```text
HostIP:8080
      ↓
10.0.0.2:80
```

### Step 2

Routing:

```text
Send through cage0
```

### Step 3

FORWARD:

```text
Allowed
```

### Step 4

Container receives:

```text
10.0.0.2:80
```

### Step 5

Web server responds.

Connection tracking remembers the DNAT mapping.

Reply automatically goes back to the client.

---

So the three concepts are:

```text
PREROUTING
    External traffic

OUTPUT
    Host-local traffic

FORWARD
    Permission to pass packet to container(this already exist)

POSTROUTING
    Masquerade rule for translating source ip (localhost/ 127.0.0.1) to bridge ip
```

Together they implement the equivalent of:

```bash
docker run -p 8080:80
```

where:

```text
Host:8080
      ↓
Container:80
```
