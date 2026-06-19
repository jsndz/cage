AppArmor and SELinux are Linux Security Modules (LSMs) that provide Mandatory Access Control (MAC).

Normally Linux security is based on users/groups (DAC - Discretionary Access Control). If a process runs as root, it can access almost everything. MAC adds another layer that can restrict even root-owned processes.

We will use AppArmor cause SElinux is little to advanced and complex to setup.

AppArmor works by attaching a security profile to a process. Every time that process asks the kernel to do something, the kernel checks the profile before allowing it.


```text
Application
     |
     | open("/etc/shadow")
     v
Kernel
     |
     | AppArmor check
     v
Profile
     |
     +--> Allow -> Success
     |
     +--> Deny  -> EPERM/EACCES
```

---

## How AppArmor Fits Into Linux

Linux has a framework called LSM (Linux Security Modules).

```text
Userspace Process
       |
       v
System Call
       |
       v
Linux Kernel
       |
       +--> DAC (user/group permissions)
       |
       +--> Capabilities
       |
       +--> AppArmor
       |
       +--> Seccomp
```

When a process calls:

```go
os.Open("/etc/shadow")
```

the kernel roughly does:

```text
1. Check filesystem permissions
2. Check capabilities
3. Check AppArmor policy
4. Return result
```

All must pass.

---

# Profiles

A profile defines what a process may do.

Example:

```text
profile cage-default {
    file,

    /usr/bin/bash rix,

    /tmp/** rw,

    deny /etc/shadow r,
}
```

Meaning:

```text
Can execute bash
Can read/write /tmp
Cannot read /etc/shadow
```

---

# How A Profile Gets Attached

When a process is executed:

```bash
/usr/bin/bash
```

AppArmor checks whether a profile exists for that executable.

```text
/usr/bin/bash
        |
        v
bash profile loaded
        |
        v
process runs confined
```

Every child process inherits confinement.

```text
bash
 └── python
      └── node
```

All remain confined unless explicitly changed.

---

# Path-Based Security

AppArmor tracks paths.

Example:

```text
deny /etc/shadow r,
allow /tmp/** rw,
```

When process does:

```go
os.Open("/etc/shadow")
```

Kernel resolves:

```text
/etc/shadow
```

Then checks rules.

```text
Match found:
deny /etc/shadow

Result:
Access denied
```

---

# Permissions

Common permissions:

```text
r  read
w  write
a  append
k  lock
m  mmap executable
l  link
```

Execution permissions:

```text
ix
px
cx
ux
```

These are important.

---

## ix (inherit execute)

```text
/usr/bin/python ix,
```

Child keeps same profile.

```text
bash(profile A)
    |
    +--> python
          |
          +--> still profile A
```

Most common for sandboxes.

---

## px (profile execute)

Switch to another profile.

```text
/usr/bin/python px,
```

```text
bash profile
      |
      +--> python profile
```

---

## ux (unconfined execute)

Dangerous.

```text
/usr/bin/python ux,
```

```text
sandbox profile
      |
      +--> python
              |
              +--> NO PROFILE
```

Avoid in Cage.

---

# Network Rules

AppArmor can restrict networking.

Example:

```text
network inet tcp,
network inet udp,
```

Allow:

```text
TCP
UDP
```

Or deny all networking.

---

# Capability Rules

Linux capabilities can also be controlled.

Example:

```text
deny capability sys_admin,
deny capability sys_module,
```

Even if process somehow has capability:

```text
CAP_SYS_ADMIN
```

AppArmor can block its usage.

---

# Mount Restrictions

Example:

```text
deny mount,
```

Blocks:

```bash
mount /dev/sda /mnt
```

Very important for containers.

---

# Signals

Can control process signaling.

Example:

```text
deny signal send,
```

Blocks:

```bash
kill -9 <pid>
```

against protected targets.

---

# Profile Modes

## Enforce

Real blocking.

```text
DENIED
```

Used in production.

---

## Complain

Logs only.

```text
ALLOWED
LOGGED
```

Useful while developing profiles.

---

# Loading Profiles

Profile file:

```text
/etc/apparmor.d/cage-default
```

Load:

```bash
sudo apparmor_parser -r /etc/apparmor.d/cage-default
```

Kernel stores compiled profile.

---

# Checking Status

```bash
sudo aa-status
```

Example:

```text
20 profiles loaded
15 in enforce mode
```

---

# How Docker Uses It

Docker starts a container.

Before executing container init process:

```text
docker-default
```

profile is attached.

```text
container init
      |
      +--> bash
      +--> node
      +--> python
```

Everything inherits confinement.

Even root inside container is restricted.

---


AppArmor is already loaded into the kernel as an LSM.
gets attached to the process is an AppArmor profile/label
Then when the process does something sensitive, the kernel invokes AppArmor's LSM hooks and passes the process profile.

   Profile      | AppArmor Behavior                | Key Restrictions
  --------------|----------------------------------|--------------------------------------------------------------------------------------------------------
    privileged  |  unconfined  — no profile loaded | Full access, no MAC restrictions
    default     | Restrictive but practical        | Blocks writes to  /proc/sys ,  /sys/firmware ,  /sys/kernel/security ; denies  mount ,  umount , 
                |                                  | pivot_root ,  ptrace ; allows all networking, general file I/O, execution
    sandbox     | Highly restrictive               | Denies all networking; read-only filesystem except  /tmp ,  /var/tmp ,  /dev/null ; denies all  mount
                |                                  | / ptrace / /proc  writes
