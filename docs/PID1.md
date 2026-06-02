These two topics become critical the moment you run untrusted processes inside Cage.

# PID 1 Behavior

Most people think PID 1 is just another process.

It isn't.

The kernel treats PID 1 specially.

---

## Normal Process Tree

```text
systemd (PID 1)
└── cage
    └── python
        └── worker
```

Suppose worker exits:

```c
exit(0);
```

The kernel does **not** immediately destroy the process.

Instead:

```text
worker
  ↓
Zombie
```

---

## What is a Zombie?

A zombie is:

```text
Process finished executing
BUT
Parent has not collected exit status
```

Kernel still stores:

```text
PID
Exit code
Accounting info
```

because parent may need it.

Example:

```c
pid = fork();

if (pid == 0) {
    exit(42);
}

sleep(100);
```

Child exits.

Parent never calls:

```c
waitpid(pid, &status, 0);
```

Result:

```text
Zombie process
```

---

## Why keep zombies?

Because parent may want:

```c
waitpid(pid, &status, 0);
```

and then:

```c
WEXITSTATUS(status)
```

to know whether child succeeded.

If kernel destroyed child immediately:

```text
Exit status lost forever
```

---

## What does waitpid() do?

```c
waitpid(child_pid, &status, 0);
```

Kernel:

```text
1. Return exit status
2. Remove zombie
3. Free process metadata
```

This is called **reaping**.

---

## Orphan Process

Suppose:

```text
Parent
 └── Child
```

Parent dies first:

```text
Parent X
```

Now:

```text
Child
```

has no parent.

This is an orphan.

Linux does:

```text
Child
   ↓
Reparent to PID 1
```

Now:

```text
PID 1
 └── Child
```

---

## Why PID 1 Matters

PID 1 becomes responsible for:

```text
Reaping orphaned children
```

If it doesn't:

```text
Zombie accumulation
```

---

## Container Problem

Suppose:

```text
Container

PID 1 = python app.py
```

Your application forks:

```text
python
 ├── worker
 ├── worker
 └── worker
```

Workers exit.

Application never calls:

```c
waitpid()
```

Now:

```text
Zombie
Zombie
Zombie
```

accumulate.

This happens surprisingly often.

---

## Signal Handling Problem

Another special thing about PID 1.

Suppose Docker wants to stop container.

It sends:

```text
SIGTERM
```

to PID 1.

If PID 1 ignores it:

```text
Container won't stop
```

Eventually:

```text
SIGKILL
```

is sent.

Application dies abruptly.

---

## Why tiny init processes exist

Tools like:

* tini
* dumb-init

exist only to be PID 1.

They:

```text
Receive signals
Forward signals
Reap zombies
```

That's their entire job.

---
