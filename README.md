```sh

mkdir -p /tmp/rootfs

docker export $(docker create ubuntu:24.04) \
  | sudo tar -C /tmp/rootfs -xf -

```

race condition
now fix code for child process ready without having network ready
using pipe

then learn about nftables and NAT 
and forwarding data