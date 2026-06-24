Rootless is a the architecture where container doesn't have access to the root permissions.
This effectively isolates the container from running any root previlages
Done with,
Creating a new user namespace
and connecting uid 0 from container to some uid in the host

This also isolates the container from having the network 
so we need alternative here will be using slirp4netns 

A rootless container runs in its own network namespace (netns).
The container has a virtual network interface (tap0).
slirp4netns runs as the same unprivileged user on the host.
It connects to the container's TAP device and implements networking in user space.
Outbound traffic is translated (NATed) and sent through the host's normal network sockets.
Replies are received by slirp4netns and forwarded back into the container.

```md 
Container
+----------------+
| eth0/tap0      |
| 10.0.2.100     |
+-------+--------+
        |
        | TAP
        |
+-------v--------+
| slirp4netns    |
| User-space NAT |
+-------+--------+
        |
        | Host sockets
        |
+-------v--------+
| Host Network   |
+----------------+
        |
     Internet
```

A TAP device is a virtual network card exposed to userspace.
The key trick is that slirp4netns uses ordinary sockets.

curl https://google.com

Step-by-step:

curl
 ↓
TCP packet
 ↓
Container eth0
 ↓
TAP device
 ↓
slirp4netns
 ↓
NAT translation
 ↓
Host TCP socket
 ↓
Internet