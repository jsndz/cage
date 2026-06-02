Clone is low level primitive of fork()
Clone has more configurations.

With clone(), you choose what is shared.
clone(flags, ...);
Flags determine what parent and child share.

Clone can be used for creating threads and containers

Linux does not have separate kernel objects for processes and threads.
Internally, both are represented by:    struct task_struct

The kernel only looks at what resources are shared.

Threads fundamentally are processes sharing Same Address Space
Share file descriptor table and share resource


Linux does:

clone(
    CLONE_VM |
    CLONE_FILES |
    CLONE_SIGHAND |
    CLONE_THREAD,
    ...
);
CLONE_VM

Share memory.

Thread A ----+
             |
             +--> Same Address Space
             |
Thread B ----+

Now:

counter++;

in Thread A is visible to Thread B immediately.

CLONE_FILES

Share file descriptor table.

Thread A
    fd 3 --> file.txt

Thread B
    fd 3 --> file.txt

Same descriptor.

CLONE_THREAD

Makes them belong to the same thread group.

From user space:

Process
 ├── Thread 1
 ├── Thread 2
 └── Thread 3

Internally Linux sees:

Task
Task
Task

sharing resources.



Container

Containers are still processes.

The trick is that they see a different environment.

Linux does:

clone(
    CLONE_NEWPID |
    CLONE_NEWNS |
    CLONE_NEWNET,
    ...
);


CLONE_NEWPID

Creates a new PID namespace.

Host:

PID 1000

Inside container:

PID 1

Same process.

Different view.

CLONE_NEWNS

Creates a new mount namespace.

Host:

/
├── home
├── var
└── etc

Container:

/
├── app
├── bin
└── lib

Different filesystem view.

CLONE_NEWNET

Creates a new network namespace.

Host:

eth0

Container:

lo

Initially only loopback.

You later attach virtual interfaces.


```md 
clone()
    creates a task

Flags determine:
    what is shared
    what is isolated

Shared memory
    => thread

New namespaces
    => container

Nothing shared
    => process

```
so thread is just another process?

At the Linux kernel level, almost yes.

A thread and a process are both represented by a task_struct.


Process
    = task with its own resources

Thread
    = task sharing resources with other tasks


So a thread is not literally a process in user-space terminology, but in the Linux kernel it is extremely close to "a process that shares almost everything with another process."


Why is PID 1 special?
1. It adopts orphaned processes

2. It reaps zombies
