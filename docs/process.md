Linux Process Handling:

A process is a running instance of a program.
When it is run the OS finds the executable file, creates process, loads to memory and start from entry point.
A process consists of:

Program code
Virtual memory
CPU registers
Threads
Open files
Environment variables
PID
Security credentials

The code is loaded and run:

Virtual Address Space

```md 
High Addresses
+------------------+
|      Stack       |
+------------------+
|                  |
|      mmap()      |
|  shared libs     |
|                  |
+------------------+
|      Heap        |
+------------------+
|      Data        |
+------------------+
|      Code        |<---- here is the code 
+------------------+
Low Addresses
```

Virtual address can be mapped to anywhere is physical memory space

The executable's code (.text) is mapped into virtual memory.
Data sections (.data, .bss) are mapped.
Shared libraries are mapped.
Stack pages are created.
Heap starts small and grows as needed.

The Pages are not ususally contigious in RAM.

Lets talk bit about process. Every process is a child of something
```md 
systemd (PID 1)
└── bash
    └── python
        └── gcc
```

So to create these children we use fork() and execve()
lets consider for  bash
when fork is run 
Creates a new task_struct
Allocate memort 
Copies process metadata.
Copies page tables.
Copies register state.
Copies file descriptor table references.
Creates parent-child relationship.

Parent PID 100
Child  PID 101

At this point,
Parent memory == Child memory
but not physically copied.
Linux uses Copy-On-Write (COW).

Copy-On-Write

Suppose parent has:

`int x = 5;`

After fork:
```md
Parent x -> Physical Page A
Child  x -> Physical Page A
```
Both point to the same RAM page.
If child changes:
`x = 10;`
Kernel sees a write to a shared page.
It creates:
```md
Parent -> Page A
Child  -> Page B
```
Now they're independent.
Without COW, fork() would be extremely expensive.


After fork returns Both processes continue from the next instruction.

Example:

pid_t pid = fork();

printf("hello\n");

Both execute:

hello
hello

because there are now two processes.

The only difference is the return value.

Parent:

pid = 101

Child:

pid = 0

Now Let's talk about execve():

Let's say the child wants to run `ls`
-->
`execve("/bin/ls", argv, envp);`

Now child :
PID is 101
Program = bash

bash code
bash heap
bash stack

Start loading /bin/ls

1. Read ELF and get 

Entry point
Required libraries
Memory layout

2. Destroy old address space

bash code
bash heap
bash stack
 is destroyed

3. Create new address space

ls code
ls data
ls heap
ls stack

4. Load shared libraries

5. Build stack
6. Reset registers
7. Jump into new program

before fork:
```md
bash (PID 100)
```

After fork:
```md
bash (PID 100)
└── bash copy (PID 101)
```
Both still running bash code.

Child execve:
```md
bash (PID 100)
└── ls (PID 101)
```


For Cage the same will happen fork and execve 
but before that there will be isolated namespace