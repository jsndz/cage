Mounting is a process of attaching a filesystem (like ext4-> real files on disk, sys-> which represents networks like eth) to a directory tree of system.

Basically Mounting the connection.

If a USB drive contains a filesystem, you can mount it at /mnt/usb:

mount /dev/sdb1 /mnt/usb

After mounting, the files on the USB appear under /mnt/usb.

Like you see when you connect the USB it automatically mounts and you can unmount it.


A mount operation:

Takes a filesystem (usually on a block device).
Attaches it to a mount point (a directory).
Makes its files accessible through the unified directory tree.


Relationship to mount namespaces

A mount namespace is a kernel feature that gives a process its own view of mounted filesystems.

Normally, all processes see the same mounts:

Process A -> /mnt/usb exists
Process B -> /mnt/usb exists

With separate mount namespaces:

Namespace 1:
  /mnt/usb mounted

Namespace 2:
  /mnt/usb not mounted

Processes in different namespaces can see different filesystem layouts.


Containers have separate mount namespace so that they have separate view of file system 


```md 
Host:
  /
  /home
  /var

Container:
  /
  /app
  /etc
```


Everything is mounted. Even / at the beginning is also mounted


There is mount table in linux that will filesystem and where are they mounted.

There are different filesystem in linux and everything should comeunder the tree mount connect them to the tree.


Why mount namespace show different view of tree?

Because the new namespace has its own mount table.

Initially, the new namespace gets a copy of the host's mount table.

when you bind the root to new FS

```sh
mount --bind /sandbox-root /
```


```md
/           SandboxRootFS
/proc       procfs
/tmp        tmpfs
```