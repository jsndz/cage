OverlayFS is probably the most important filesystem concept behind Docker.

Before OverlayFS, ask:

> If 100 containers use the same Ubuntu image, do we store 100 copies of Ubuntu?

That would be wasteful.

---

# The Problem

Suppose you have:

```text
ubuntu-rootfs/

├── bin
├── etc
└── usr
```

and:

```text
etc/passwd
```

contains:

```text
root:x:0:0
```

Container A starts.

Container B starts.

Both use the same rootfs.

---

Now Container A edits:

```text
/etc/passwd
```

If we directly modify:

```text
ubuntu-rootfs/etc/passwd
```

then:

```text
Container B
```

would also see the change.

Bad.

Containers must have isolated writes.

---

# Naive Solution

Copy entire rootfs.

```text
ubuntu-rootfs
        ↓
copy
        ↓
container-rootfs
```

Now modify:

```text
container-rootfs/etc/passwd
```

Works.

But:

```text
Ubuntu image = 500MB

100 containers
=
50GB
```

Wasteful.

---

# OverlayFS Solution

Split filesystem into layers.

```text
Lower Layer
    =
Original Ubuntu image
    =
Readonly
```

```text
Upper Layer
    =
Container writes
```

```text
Work Directory
    =
Internal OverlayFS bookkeeping
```

```text
Merged Directory
    =
What process sees
```

---

Visual:

```text
Lower
├── bin
├── etc
└── usr

Upper
(empty)

Merged
```

Process uses:

```text
Merged
```

not lower or upper directly.

---

# Mounting OverlayFS

```bash
mount -t overlay overlay \
-o lowerdir=lower,upperdir=upper,workdir=work \
merged
```

Kernel creates:

```text
Merged View
```

---

# Read Path

Suppose:

```text
lower/etc/passwd
```

exists.

Upper is empty.

Process reads:

```text
merged/etc/passwd
```

Kernel:

```text
Check Upper
      ↓
Not Found
      ↓
Check Lower
      ↓
Found
```

Returns lower file.

---

Visual:

```text
merged/etc/passwd
         ↓
upper/etc/passwd ?
         ↓
No
         ↓
lower/etc/passwd
```

---

# First Write

Suppose process does:

```bash
echo hello >> merged/etc/passwd
```

Problem:

```text
Lower layer is readonly
```

Can't modify it.

---

Kernel performs:

## Copy-Up

Copies file:

```text
lower/etc/passwd
```

to:

```text
upper/etc/passwd
```

Now:

```text
upper/etc/passwd
```

exists.

Then modification happens there.

---

Result:

```text
Lower

etc/passwd
    root:x:0:0
```

unchanged.

---

```text
Upper

etc/passwd
    root:x:0:0
    hello
```

modified.

---

# Future Reads

Now process reads:

```text
merged/etc/passwd
```

Kernel:

```text
Check Upper
      ↓
Found
```

Uses upper version.

Lower version becomes hidden.

---

Think:

```text
merged/etc/passwd
         ↓
upper/etc/passwd
```

---

# Copy-on-Write

This is the same idea as process memory COW.

Memory COW:

```text
Parent Page
       ↓
Shared
       ↓
Write
       ↓
Copy
```

OverlayFS:

```text
Lower File
      ↓
Read Shared
      ↓
Write
      ↓
Copy-Up
      ↓
Modify Copy
```

Hence:

```text
Copy-On-Write
```

---

# File Creation

Suppose:

```bash
touch merged/newfile
```

No file exists below.

Kernel simply creates:

```text
upper/newfile
```

---

Result:

```text
Upper

newfile
```

---

# File Deletion

Interesting case.

Suppose:

```text
lower/file.txt
```

exists.

Process:

```bash
rm merged/file.txt
```

Kernel cannot remove:

```text
lower/file.txt
```

because lower is readonly.

Instead it creates a special marker in upper.

Conceptually:

```text
upper/.wh.file.txt
```

(whiteout file)

Meaning:

```text
Hide lower/file.txt
```

---

Now reads behave as if file is deleted.

Lower still contains it.

---

# Why Docker Loves OverlayFS

Suppose image:

```text
Ubuntu
```

stored once:

```text
Lower Layer
```

Container A:

```text
Upper A
```

Container B:

```text
Upper B
```

Container C:

```text
Upper C
```

Visual:

```text
             Ubuntu Image
                   ↓
                Lower

      ┌────────┼────────┐

    UpperA   UpperB   UpperC

      ↓         ↓        ↓

    MergedA  MergedB  MergedC
```

All containers share Ubuntu.

Only changes consume space.

---

# What the Container Sees

Container never sees:

```text
lower
upper
work
```

It only sees:

```text
merged

/
├── bin
├── etc
├── usr
└── home
```

Looks like a normal filesystem.

---

# Important Exercise

Create:

```bash
mkdir lower upper work merged

echo hello > lower/file.txt
```

Mount overlay:

```bash
sudo mount -t overlay overlay \
-o lowerdir=lower,upperdir=upper,workdir=work \
merged
```

Check:

```bash
cat merged/file.txt
```

Output:

```text
hello
```

Now modify:

```bash
echo world >> merged/file.txt
```

Inspect:

```bash
find upper
```

You'll see:

```text
upper/file.txt
```

appeared.

That is the copy-up operation happening in real time.

---

# Why This Matters For Cage

Without OverlayFS:

```text
Container
     ↓
Copies rootfs
```

Huge storage waste.

With OverlayFS:

```text
Ubuntu Image (readonly)
          +
Container Write Layer
          ↓
Merged View
```

Every container gets:

```text
/
├── bin
├── etc
├── usr
```

without copying the image.

That's essentially the filesystem model used by Docker, containerd, and most OCI runtimes today. OverlayFS is what makes containers lightweight instead of miniature virtual machines.



Yes.

OverlayFS is a filesystem implemented inside the Linux kernel.

Just like:

```text
ext4
xfs
tmpfs
procfs
```

are filesystem drivers, Linux also has:

```text
overlayfs
```

built into the kernel (or as a kernel module).

---

When you run:

```bash
sudo mount -t overlay overlay \
-o lowerdir=lower,upperdir=upper,workdir=work \
merged
```

you're telling the kernel:

```text
Create an OverlayFS mount.

Lower Layer = lower
Upper Layer = upper
Work Dir    = work

Expose result at:

merged
```

The kernel creates a new mount object:

```text
merged
    ↓
OverlayFS
```

---

Think about what happens when a process does:

```bash
cat merged/file.txt
```

The process doesn't know OverlayFS exists.

It simply asks:

```c
open("merged/file.txt")
```

The VFS layer sees:

```text
merged
    ↓
OverlayFS mount
```

and forwards the operation to the OverlayFS driver.

---

OverlayFS code inside the kernel then does:

```text
Does upper/file.txt exist?
```

If yes:

```text
return upper/file.txt
```

Otherwise:

```text
return lower/file.txt
```

---

Conceptually:

```text
cat merged/file.txt

        ↓

     OverlayFS

        ↓

upper/file.txt ?
      /   \
    yes    no
    /       \
upper      lower
```

---

The same thing happens for writes.

Suppose:

```bash
echo hello >> merged/file.txt
```

OverlayFS receives:

```text
write file.txt
```

It checks:

```text
Is file.txt already in upper?
```

If not:

```text
Copy lower/file.txt
        ↓
upper/file.txt
```

Then:

```text
Write to upper/file.txt
```

This is the "copy-up" operation.

---

The important realization is:

```text
lower
upper
work
```

are just ordinary directories.

Nothing special.

Example:

```text
lower/
└── file.txt

upper/
(empty)

work/
(empty)
```

The magic comes from:

```text
OverlayFS kernel driver
```

which merges them into:

```text
merged/
└── file.txt
```

---

A useful analogy is:

```text
ext4
    =
filesystem backed by disk

tmpfs
    =
filesystem backed by RAM

overlayfs
    =
filesystem backed by other filesystems
```

OverlayFS doesn't store data itself.

Instead it combines existing filesystems into a new virtual filesystem view.

---

Internally, the flow is roughly:

```text
Process
    ↓
open("/merged/file.txt")
    ↓
VFS
    ↓
OverlayFS driver
    ↓
Upper/Lower lookup
    ↓
Real filesystem (ext4/xfs/etc.)
```

So yes, OverlayFS is a kernel-provided filesystem implementation, and the `mount -t overlay ...` command is simply telling the kernel to create an OverlayFS mount using those directories as its backing layers.
