

Seccomp (Secure Computing Mode) is a Linux kernel feature that restricts which **syscalls** a process can make.

Think of it as a firewall for syscalls.

Without seccomp:

```text
Container Process
       |
       v
Any Linux syscall
(open, mount, ptrace, reboot, clone, ...)
```

With seccomp:

```text
Container Process
       |
       v
Seccomp Filter
       |
       +--> allowed syscalls
       |
       +--> blocked syscalls
```

Why it matters for Cage:

Namespaces isolate resources.
Cgroups limit resources.

But a process can still ask the kernel to do dangerous things through syscalls.

Example:

```c
mount(...)
ptrace(...)
reboot(...)
kexec_load(...)
```

Seccomp can block these even if the process is compromised.

Example policy:

```text
ALLOW:
read
write
exit
futex
mmap
munmap
openat
close

DENY:
mount
umount2
ptrace
kexec_load
reboot
swapon
swapoff
```

When a blocked syscall is executed:

```text
Process
   |
   +--> mount(...)
            |
            v
        EPERM
```

or the kernel can kill the process:

```text
SCMP_ACT_KILL
```

Common actions:

| Action | Meaning |
|----------|----------|
| ALLOW | Permit syscall |
| ERRNO | Return error (EPERM) |
| KILL | Kill process |
| LOG | Log syscall |
| TRACE | Notify tracer |

Example with libseccomp:

```go
filter, _ := seccomp.NewFilter(seccomp.ActErrno.SetReturnCode(int16(syscall.EPERM)))

filter.AddRule(seccomp.Syscall(syscall.SYS_READ), seccomp.ActAllow)
filter.AddRule(seccomp.Syscall(syscall.SYS_WRITE), seccomp.ActAllow)
filter.AddRule(seccomp.Syscall(syscall.SYS_EXIT), seccomp.ActAllow)

filter.Load()
```

For Cage, a typical hardening progression is:

1. Namespaces
2. Cgroups
3. Capabilities drop
4. `NO_NEW_PRIVS`
5. Read-only rootfs
6. Seccomp
7. AppArmor/SELinux
8. User namespaces

Docker follows a similar approach and ships with a default seccomp profile that blocks dozens of dangerous syscalls while allowing normal applications to run.

For Cage, start with a **default-deny profile**:

```text
Default: ERRNO

Allow:
read
write
openat
close
mmap
munmap
brk
futex
clock_gettime
rt_sigaction
rt_sigprocmask
exit
exit_group
```

Then run programs and keep adding required syscalls until they work. This is how most container runtimes build hardened seccomp policies.