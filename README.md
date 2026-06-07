```sh

mkdir -p /tmp/rootfs

docker export $(docker create ubuntu:24.04) \
  | sudo tar -C /tmp/rootfs -xf -

```