A namespace is a Linux kernel feature that isolates a specific type of resource and gives processes their own view of that resource.
So lets say you have set of resources like time, pid or network.
if you want a process to have its own isolated view of the network, processes, filesystem mounts, etc., you place it in separate namespaces for those resources.
You can say that this process can only see these pid or network. 

Types of Namespaces in linux:
mnt,pid,net,ipc,uts,cgroup,user,time


so basically there are various resource 
using namespace you make it such that a process can only see select few?

A namespace doesn't usually let a process see only "select few resources." It lets a process see a separate instance of a resource type.
"This process can only see these few PIDs" X
"This process sees its own PID namespace, which contains only the processes that belong to that namespace." 

How to make a process join new space:

1. Create a new process in a new namespace

```bash
sudo unshare --pid --fork bash
```

`unshare` creates a new PID namespace, and `bash` starts inside it. Any child processes of that shell also belong to that namespace.

2. Create a namespace and run a command in it

```bash
sudo ip netns add ns1
sudo ip netns exec ns1 bash
```

The `bash` process runs inside the network namespace `ns1`.

3. Join an existing namespace

Find a process already in the target namespace:

```bash
ps aux
```

Then join its namespaces:

```bash
sudo nsenter --target <PID> --net --pid --mount bash
```

This starts a shell in the same namespaces as that process.

Under the hood, the kernel provides system calls:

* `clone()` → create a process in new namespaces.
* `unshare()` → move the current process into new namespaces.
* `setns()` → join an existing namespace.


The easiest example is a **PID namespace**.

On the host machine:

```bash
ps
```

Output:

```text
PID  COMMAND
1    systemd
100  nginx
200  mysql
500  bash   
```

Now create a new PID namespace:

```bash
sudo unshare --pid --fork bash
```

Inside that shell:

```bash
ps
```

Output:

```text
PID  COMMAND
1    bash
2    ps
```

What happened?

* The host still has hundreds of processes.
* The new shell cannot see them.
* Inside the namespace, `bash` thinks it is PID 1.
* It sees only processes that belong to its PID namespace.

This is the essence of namespaces:

> The kernel gives the process a different view of a resource. The real system hasn't changed; only what the process can see has changed.



Each namespace creates a different kernel structure
There is an initial namespace for every namespace type, often called the root or initial namespace.\
When Linux boots:

Kernel

creates:

Initial PID Namespace -> this has some DS
Initial Mount Namespace -> same here with mount table 
Initial Network Namespace
Initial UTS Namespace

so when you create a new namespace you are essentially creating a new data structure for that specific ns

A process is associated with a namespace through pointers stored in its task_struct.

Very roughly:

Process
   |
task_struct
   |
nsproxy
   |
   ├── mnt_ns
   ├── pid_ns
   ├── net_ns
   └── uts_ns

child inherits the namespace from parent

Run:

ls -l /proc/self/ns

You see:

mnt:[4026532902]
pid:[4026532901]

Those are basically handles to the namespace objects.

The kernel is exposing:

Current Process
       |
       +--> Namespace Object

through procfs.