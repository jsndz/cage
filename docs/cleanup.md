for cleaning up the mounts:
You don't need to clean:

/oldroot
/proc
/dev
/sys

after:

execve("/bin/bash")

because when the container process exits:

Mount namespace destroyed
        ↓
all namespace-private mounts disappear

Kernel cleans them automatically.


Unmount tmp/overlay/merged

and remove the tmp/overlay folder

and finally delete the cgroup by deleting the cgroup/cage folder