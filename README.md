```sh

mkdir -p /tmp/rootfs

docker export $(docker create alpine:latest) \
  | sudo tar -C /tmp/rootfs -xf -

```

