Diving the capabilities of linux root into different things.


example:
| Capability     | Meaning              |
| -------------- | -------------------- |
| CAP_NET_ADMIN  | configure interfaces |
| CAP_SYS_ADMIN  | almost god mode      |
| CAP_SYS_MODULE | load kernel modules  |
| CAP_SYS_PTRACE | inspect processes    |
| CAP_SYS_TIME   | change system clock  |

Container if run as root may perform dangerous things so to stop that we use capablilites.


Linux capalities is the solution which helps in Keeping only what is required.

i.e, split the power of root into smaller privileges

Initial model of linux was that uid =0 is all powerful and access everything

but with LC A process can have some privileges without having all of them.

Process Capability Sets

A process actually has multiple capability sets.

Permitted
Effective
Inheritable
Bounding
Ambient

The most important are:

1. Effective Set
Capabilities currently active.
The kernel checks this set. This indicates enabled set of capabilities

2. Permitted Set
Maximum capabilities process may enable.
These are set of capaablities that process can have.


3. Bounding Set
If a capability is removed from the Bounding Set: Process can never obtain it again.

4. Ambient Set
Capabilities preserved across exec.


Most Important Capabilities
CAP_SYS_ADMIN: Same as root
CAP_NET_ADMIN: Network administration.
CAP_NET_RAW: Raw sockets.
CAP_SYS_PTRACE: Trace memory and inspect processes.
CAP_SYS_MODULE: Load kernel modules.
CAP_SYS_TIME: Change system clock.
CAP_SYS_BOOT: Allows reboot
CAP_CHOWN: Allows changing file ownership

How does kernel check it?

for any syscall the kernel calls capable function which checks in task struct

task_struct
    ├── PID
    ├── memory info
    ├── open files
    ├── credentials
    └── capabilities


Remove all the caps from bounding  set
Add the caps you want to permitted and effective set
set PR_SET_NO_NEW_PRIVS so that it stops process from gaining privileges it didn't already have

// Clear all capability sets: CAPS covers EFFECTIVE, PERMITTED, and INHERITABLE.
	// Clearing INHERITABLE prevents capability leakage across exec() calls.
	// AMBIENT caps require the cap to be in both PERMITTED and INHERITABLE,
	// so leaving INHERITABLE empty guarantees AMBIENT is also empty.