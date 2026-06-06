1. Minimal Container Runtime

Before building:
Learn:

* namespaces
* process isolation
* mount namespaces
* PID namespaces

Build:

```bash
cage run /bin/bash 
# go run cmd/main.go sh
```

Implement:

* `clone()` / `unshare()`
* PID namespace
* mount namespace
* hostname namespace
* basic filesystem isolation

Goal:

```text id="fr8efj"
Run a process isolated from the host.
```

So basically create a isolated namespace then run clone and process and do execvc
2 ways to do it 

1. create a child and then isolate it using unshare
unshare() operates on the calling process. You can call it directly in the current process:
However, there are namespace-specific caveats:
For most namespaces (CLONE_NEWNS, CLONE_NEWUTS, CLONE_NEWNET, etc.), unshare() immediately places the calling process into the new namespace.
For a PID namespace (CLONE_NEWPID), the calling process does not move into the new PID namespace. Instead, the next child created after unshare(CLONE_NEWPID) becomes PID 1 in the new PID namespace.
2. directly create a isolated namespace with clone
Better use this

for this we can run syscall clone directly 
we will exec the command with clone flags