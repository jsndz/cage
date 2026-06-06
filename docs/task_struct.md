Excellent question. You're now digging into actual kernel internals.

## What is `task_struct`?

`task_struct` is the kernel's representation of a running task.

Very roughly:

```c
struct task_struct {
    pid;
    state;
    mm;
    files;
    nsproxy;
    ...
};
```

When you think:

```text
Process
```

the kernel thinks:

```text
task_struct
```

---

## Where is it stored?

In kernel memory (RAM).

Not on disk.

Not in `/proc`.

When you run:

```bash
sleep 100
```

the kernel allocates a `task_struct` in kernel memory.

Conceptually:

```text
Kernel Memory

task_struct (sleep)
task_struct (bash)
task_struct (systemd)
...
```

---

## What does it contain?

A lot of things.

Conceptually:

```text
task_struct

├── PID
├── State
├── CPU Scheduling Info
├── Memory Info
├── Open Files
├── Namespace References
├── Parent Process
├── Child Processes
├── Signal Information
└── Credentials
```

---

### Parent/Child

```text
bash
 └── sleep
```

Kernel stores:

```text
sleep.task_struct
    |
    +--> parent = bash.task_struct
```

---

### Memory

```text
task_struct
     |
     +--> mm_struct
```

`mm_struct` describes:

```text
Virtual Address Space

Code
Heap
Stack
Shared Libraries
```

---

### Files

```text
task_struct
     |
     +--> files_struct
```

Contains:

```text
fd 0
fd 1
fd 2
fd 3
...
```

Open file descriptors.

---

### Namespaces

```text
task_struct
     |
     +--> nsproxy
```

Which points to:

```text
Mount Namespace
PID Namespace
Network Namespace
...
```

---

## Is task_struct only for processes?

Here's the interesting part.

Linux doesn't really distinguish processes and threads internally.

Both are:

```text
task_struct
```

Example:

```bash
Process
```

↓

```text
task_struct
```

Thread:

```text
task_struct
```

also.

---

This is why threads are created with:

```c
clone()
```

They are just tasks that share more resources.

---

## What About Files?

Files do NOT have a `task_struct`.

Files have completely different kernel structures.

Conceptually:

```text
File on disk
     |
     +--> inode
```

Open file:

```text
Open File Object
```

Directory entry:

```text
dentry
```

---

Very roughly:

```text
Process
   |
task_struct
   |
files_struct
   |
file
   |
inode
```

---

Example

Suppose:

```bash
cat file.txt
```

Kernel objects:

```text
task_struct (cat process)
     |
files_struct
     |
file object
     |
inode
```

Each layer represents something different.

---

## Relationship Between Major Objects

For Cage, think:

```text
Running Process
      ↓
task_struct
```

```text
Memory
      ↓
mm_struct
```

```text
Namespaces
      ↓
nsproxy
```

```text
Open Files
      ↓
files_struct
```

```text
File on Disk
      ↓
inode
```

---

A useful mental model:

```text
Userspace Process
        ↓
Kernel Representation
        ↓
task_struct
```

Whenever you run:

```bash
sleep 100
```

or:

```bash
python app.py
```

or:

```bash
nginx
```

the kernel creates a `task_struct`.

Almost everything you've learned so far—PIDs, namespaces, signals, scheduling, parent/child relationships—ultimately hangs off that `task_struct`. That's why it's one of the most important structures in the Linux kernel.
