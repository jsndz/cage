  ### The Correct Lifecycle Sequence
    sequenceDiagram                                                                                                                                         
        participant Parent as Parent (Host)                                                                                                                 
        participant Kernel as Linux Kernel                                                                                                                  
        participant Child as Child (Container Init)                                                                                                         
                                                                                                                                                            
        Parent->>Kernel: Create cgroup & write limits (memory.max, cpu.max, pids.max)                                                                       
        Parent->>Child: Spawn child process (cmd.Start)                                                                                                     
        Note over Child: Child blocks on pipe read<br/>(waiting for Parent signal)                                                                          
        Parent->>Kernel: Write Child PID to cgroup.procs                                                                                                    
        Parent->>Parent: Configure namespaces & network                                                                                                     
        Parent->>Child: Write to pipe (unblock Child)                                                                                                       
        Note over Child: Child reads pipe                                                                                                                   
        Child->>Kernel: Exec user command (execve)                                                                                                          

  1. Step 1: Create cgroup and set limits (Parent)
  Before spawning the process, the parent creates the directory  /sys/fs/cgroup/cage/<container_id>  and writes the resource limits (like  memory.max ).
  2. Step 2: Start the process in a suspended state (Parent & Child)
  The parent starts the child process (e.g. via  cmd.Start() ). The child begins executing the container setup but immediately blocks (e.g., reading from a
  sync pipe). It does not yet execute the untrusted target application (like  /bin/sh  or python scripts).
  3. Step 3: Move the process into the cgroup (Parent)
  The parent reads the child's  PID , and writes it to  /sys/fs/cgroup/cage/<container_id>/cgroup.procs .
  4. Step 4: Unblock the child (Parent & Child)
  Now that the child is safely isolated in its cgroup, the parent writes to the sync pipe to unblock the child.
  5. Step 5: Execute the user workload (Child)
  The child unblocks and calls  syscall.Exec("/bin/sh")  (or the requested workload). Because it is already in the cgroup, the limits are applied to the
  workload from the very first CPU cycle.