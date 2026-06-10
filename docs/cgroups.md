cgroups tells you how much resource can a process get.
cgroups (Control Groups) are a Linux kernel feature for organizing processes into groups and controlling, accounting, and isolating their resource usage.

Creating a Cgroup
mkdir /sys/fs/cgroup/mygroup

Kernel creates a new cgroup.

Move a process into it:

echo 1234 > /sys/fs/cgroup/mygroup/cgroup.procs

Now PID 1234 belongs to that group.

CPU Controller

Controls CPU scheduling.

Limit to one CPU
echo "100000 100000" > cpu.max

Meaning:

quota = 100000 us
period = 100000 us

So:

100%

of one CPU.

At a high level, a cgroup is just a kernel-managed object that sits between processes and resource allocators.

Without cgroups:

```md
Process
│
▼
Kernel scheduler
Memory allocator
IO subsystem
```

With cgroups:

```md
Process
│
▼
Cgroup
│
▼
Kernel resource controllers
```

Internally

When you create:

mkdir /sys/fs/cgroup/mygroup

you are not creating a normal directory.

The cgroup filesystem (cgroupfs) receives the request and asks the kernel to create a new internal cgroup structure.

Attaching a process

When you do:

echo 1234 > cgroup.procs

the kernel:

Finds PID 1234.
Finds the cgroup.
Updates the task's cgroup membership.

Conceptually:

task->cgroup = mygroup;

Every thread has references to its cgroup state.

CPU example

Suppose:

echo "50000 100000" > cpu.max

Meaning:

50ms quota
100ms period

The kernel stores:

cgroup->cpu.quota = 50000;
cgroup->cpu.period = 100000;

During scheduling

Linux uses the Completely Fair Scheduler (CFS).

Normally CFS picks the next runnable process.

With cgroups it does:

Pick task
│
▼
Check task's cgroup
│
▼
Does cgroup still have CPU budget?

If yes:

Run task

If no:

Throttle task

until the next period.

Budget accounting

Imagine:

Quota = 50ms
Period = 100ms

Process runs:

10ms
20ms
15ms

Total:

45ms

Still allowed.

Runs another:

10ms

Now:

55ms used

Kernel marks the cgroup:

THROTTLED

No process inside that cgroup gets CPU until the next period starts.


Why is it a hierarchy?

Cgroups form a tree:

root
 ├── A
 │   ├── A1
 │   └── A2
 └── B

Suppose:

A = 4 CPUs

Then:

A1 + A2 <= 4 CPUs

The kernel propagates limits down the tree.

This is why Kubernetes can create:

system
 └── pod
      └── container



for containers you ahve create a cgroup outside the ccontainer like in the main kernel