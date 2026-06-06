Now you're reaching the transition point between **mount namespaces** and **real containers**.

`chroot()` changes what a process considers to be `/`.

---

## Normal Process

Suppose the host filesystem is:

```text
/
├── bin
├── etc
├── home
└── var
```

Process does:

```c
open("/etc/passwd");
```

Kernel resolves:

```text
/
 └── etc
      └── passwd
```

from the host root.

---

## chroot()

Suppose you create:

```text
/sandbox
├── bin
├── etc
└── lib
```

Then:

```bash
chroot /sandbox /bin/bash
```

or internally:

```c
chroot("/sandbox");
```

The kernel changes the process's root directory.

Before:

```text
Process Root
      |
      +--> /
```

After:

```text
Process Root
      |
      +--> /sandbox
```

---

Now when the process opens:

```c
open("/etc/passwd");
```

the kernel interprets it as:

```text
/sandbox/etc/passwd
```

not:

```text
/etc/passwd
```

---

## What Actually Changes?

Inside `task_struct` there is filesystem state.

Very roughly:

```text
task_struct
    |
    +--> fs_struct
```

which contains:

```text
Root Directory
Current Working Directory
```

Conceptually:

```text
fs_struct

root = /
cwd  = /home/jaison
```

After:

```c
chroot("/sandbox");
```

becomes:

```text
fs_struct

root = /sandbox
cwd  = ...
```

Now path resolution starts from `/sandbox`.

---

## Example

Host:

```text
/
├── etc
│   └── passwd
└── sandbox
    └── etc
        └── passwd
```

After:

```bash
chroot /sandbox
```

Inside:

```bash
cat /etc/passwd
```

actually reads:

```text
/sandbox/etc/passwd
```

on the host.

---

## Why chroot Is Not A Container

Many beginners think:

```text
chroot == container
```

Not true.

`chroot()` only changes path resolution.

It does NOT create:

* PID namespace
* Mount namespace
* Network namespace
* User namespace

Example:

Inside chroot:

```bash
ps aux
```

still shows host processes.

Because PID namespace hasn't changed.

---

## Classic Problem

Suppose inside chroot:

```text
/
├── bin
└── etc
```

Process still has an open file descriptor pointing outside.

Or has root privileges.

It may escape.

Historically, `chroot` was never intended as a security boundary.

---

## Why Containers Use Mount Namespaces Too

Instead of:

```text
Host
└── /sandbox
```

containers typically create a new mount namespace first.

Then:

```text
Mount Namespace

/ -> ContainerRootFS
```

Now the process genuinely sees:

```text
/
├── bin
├── etc
├── lib
└── proc
```

as its filesystem.

The host's root is no longer visible through the mount tree.

---

## How Path Resolution Changes

Normal:

```c
open("/etc/passwd");
```

Kernel:

```text
root = /
      ↓
/etc/passwd
```

After chroot:

```text
root = /sandbox
      ↓
/sandbox/etc/passwd
```

Everything beginning with:

```text
/
```

is interpreted relative to the new root.

---

## How This Fits Into Cage

A typical container setup is:

```text
clone(CLONE_NEWNS)
        ↓
new mount namespace
        ↓
mount container rootfs
        ↓
chroot() or pivot_root()
        ↓
execve("/bin/bash")
```

Result:

```text
Container

/
├── bin
├── etc
├── lib
└── proc
```

The process believes that filesystem is the entire machine.

---

### Simple Mental Model

```text
Mount Namespace
    =
Which filesystem tree exists?

chroot()
    =
Where does "/" point inside that tree?
```

Example:

```text
Filesystem Tree

/host-root
└── containers
     └── cage-rootfs
```

Mount namespace decides what tree is visible.

`chroot("/host-root/containers/cage-rootfs")` decides that this directory becomes `/` for that process. That's why `chroot` is often called "change root".


chroot does change the root dir but it does not remove old root from view 

we use pivot root for changing the mount also

pivot_root() actually changes the mount tree.

Suppose you're inside a new mount namespace.

Current:

/
├── home
├── etc
├── var
└── newroot
    ├── bin
    ├── etc
    └── lib

Run:

pivot_root("/newroot", "/newroot/oldroot");

The kernel swaps roots.

After:

/
├── bin
├── etc
├── lib
└── oldroot
    ├── home
    ├── etc
    └── var

Now the new root filesystem is actually mounted at /.

The old host root gets moved under /oldroot.


Then container runtimes do:

umount /oldroot

Result:

/
├── bin
├── etc
└── lib

Host root is gone.

Not merely hidden.

Disconnected from the container's mount tree.


in simple words
you create new root 
you create a folder called /newroot/oldroot then 
then you swap them 
pivot change root pointing from -> / to -> /newroot
and -> / this is pointed from /oldroot
then you can unmount /oldroot