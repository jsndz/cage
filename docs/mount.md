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


Lets take the prev example of USB 

when you connect usb -> mount /dev/sdb1 /mnt
the /mnt is a mount point which is nothing but the directory

kernel will be like 

Filesystem on /dev/sdb1
        ↓
show it at
        ↓
/mnt


Suppose the USB contains:

USB Filesystem

photos/
movie.mp4
notes.txt

Before mounting:

/
├── home
├── tmp
└── mnt

After:

/
├── home
├── tmp
└── mnt
    ├── photos
    ├── movie.mp4
    └── notes.txt


What Really Happened?

The kernel stores something conceptually like:

Mount Table

Source      Target

/dev/sda1   /
/dev/sdb1   /mnt
procfs      /proc

When a process accesses:

/mnt/file1

the kernel checks:

Which filesystem is mounted at /mnt?

Answer:

/ dev / sdb1

Then reads from that filesystem.

Its not copying , The kernel is simply changing where path resolution goes.
Like redirecting to the files

This is put in mount table
And when you do unmount the entry in table is removed

mount --bind uses same mechanism 

src -> destination

Everytime kernel consults mount information during path resolution.

mount table is stored inside kernel.


To understand Mount lets run some commands:


make a playground :


```sh
mkdir -p ~/mount-lab
cd ~/mount-lab

mkdir src dst
echo "hello from src" > src/test.txt

```
Tree:
```md 
pop-os:~/mount-lab$ tree
.
├── dst
└── src
    └── test.txt


```
A normal mount attaches a filesystem to a directory.
mount -t ext4 /dev/sda1 /mnt
Disk partition (/dev/sda1)
        ↓
     ext4 driver
        ↓
       /mnt

The kernel reads the filesystem metadata and exposes its contents at /mnt.

Or:

mount -t tmpfs tmpfs /tmp

creates a brand-new in-memory filesystem.

A bind mount attaches an existing path to another path.

mount --bind /home/jaison/project /container/app
/home/jaison/project
           ↓
     same files
           ↓
/container/app

No new filesystem is mounted.

No filesystem driver is involved.

No disk is read.

The kernel just creates another mount point referring to the same underlying files.

Lets bind mount 


sudo mount --bind src dst

this connects dst to src

Now:

ls dst
cat dst/test.txt

Output:

test.txt
hello from src

but the dst did not have any test.txt

it did not copy or move.

mount table will have dst -> src mapping 

so when we do bind it creates something like a mapping so when you go to the some dir during path resolution 
you will be redirected to the  specific memory


jaison@pop-os:~/mount-lab$ cat /proc/self/mounts | grep mount-lab
jaison@pop-os:~/mount-lab$ sudo mount --bind src dst
jaison@pop-os:~/mount-lab$ cat /proc/self/mounts | grep mount-lab
/dev/nvme0n1p7 /home/jaison/mount-lab/dst ext4 rw,noatime,errors=remount-ro 0 0
jaison@pop-os:~/mount-lab$ sudo umount dst
jaison@pop-os:~/mount-lab$ cat /proc/self/mounts | grep mount-lab
jaison@pop-os:~/mount-lab$ 


here as you can see mount table was created when you add something to mount


A mount tells VFS:

When you reach this directory,
switch to another filesystem root.

Example:

/
├── home
├── proc
└── tmp

Mount table:

/proc -> procfs
/tmp  -> tmpfs

When path resolution reaches /proc, VFS switches into the proc filesystem.


lets say 

open("/home/jaison/file.txt");

Start at root
/
Lookup "home"
dentry("home")
      ↓
inode
Lookup "jaison"
dentry("jaison")
      ↓
inode
Lookup "file.txt"
dentry("file.txt")
      ↓
inode

Now VFS has the inode of the file.

Then the filesystem driver reads the actual data.

lets say

/
├── home
└── proc


lookup proc 

and notices the mount there 
now rather than calling the syscall for ext4 
the vfs calls syscall for procfs root


