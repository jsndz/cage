Proc is a Virtual file system that is generated dynamically by the kernel.
It does not exist in the file.

cat /proc/cpuinfo

There is no actual cpuinfo file on disk. The kernel creates the content on demand.

Linux follows the philosophy:

"Everything is a file."

Instead of creating special APIs for process information, the kernel exposes information through files.

/proc/PID

Every process gets a directory.
ls /proc will give you the list of PIDs 
here PIDs have info that is generated dynamically by the terminal.


Proc itself is a file system that holds process information and represented as a file. (procfs)
/proc consults the PID namespace of the reader.


/proc/self -> gives the info about current process

the kernel does not read a file from disk. Instead, the procfs code looks at your process's internal kernel structures (task_struct, memory mappings, namespaces, credentials, etc.) and generates the text on demand.
The data already exists inside kernel structures. Procfs is mostly a formatting layer.
It is not stored or written into memory.

cat /proc/self/status is important to get the status and info of the running process


We can use this also for identifing which namespaces does process in.


Terminal 1:
jaison@pop-os:~$ ls -l /proc/self/ns
total 0
lrwxrwxrwx 1 jaison jaison 0 Jun  2 16:16 cgroup -> 'cgroup:[4026531835]'
lrwxrwxrwx 1 jaison jaison 0 Jun  2 16:16 ipc -> 'ipc:[4026531839]'
lrwxrwxrwx 1 jaison jaison 0 Jun  2 16:16 mnt -> 'mnt:[4026531832]'
lrwxrwxrwx 1 jaison jaison 0 Jun  2 16:16 net -> 'net:[4026531833]'
lrwxrwxrwx 1 jaison jaison 0 Jun  2 16:16 pid -> 'pid:[4026531836]'
lrwxrwxrwx 1 jaison jaison 0 Jun  2 16:16 pid_for_children -> 'pid:[4026531836]'
lrwxrwxrwx 1 jaison jaison 0 Jun  2 16:16 time -> 'time:[4026531834]'
lrwxrwxrwx 1 jaison jaison 0 Jun  2 16:16 time_for_children -> 'time:[4026531834]'
lrwxrwxrwx 1 jaison jaison 0 Jun  2 16:16 user -> 'user:[4026531837]'
lrwxrwxrwx 1 jaison jaison 0 Jun  2 16:16 uts -> 'uts:[4026531838]'
jaison@pop-os:~$ 


Terminal 2:
jaison@pop-os:~$ unshare --uts bash
unshare: unshare failed: Operation not permitted
jaison@pop-os:~$ sudo unshare --uts bash
[sudo] password for jaison: 
root@pop-os:/home/jaison# ls -l /proc/self/ns
total 0
lrwxrwxrwx 1 root root 0 Jun  2 16:17 cgroup -> 'cgroup:[4026531835]'
lrwxrwxrwx 1 root root 0 Jun  2 16:17 ipc -> 'ipc:[4026531839]'
lrwxrwxrwx 1 root root 0 Jun  2 16:17 mnt -> 'mnt:[4026531832]'
lrwxrwxrwx 1 root root 0 Jun  2 16:17 net -> 'net:[4026531833]'
lrwxrwxrwx 1 root root 0 Jun  2 16:17 pid -> 'pid:[4026531836]'
lrwxrwxrwx 1 root root 0 Jun  2 16:17 pid_for_children -> 'pid:[4026531836]'
lrwxrwxrwx 1 root root 0 Jun  2 16:17 time -> 'time:[4026531834]'
lrwxrwxrwx 1 root root 0 Jun  2 16:17 time_for_children -> 'time:[4026531834]'
lrwxrwxrwx 1 root root 0 Jun  2 16:17 user -> 'user:[4026531837]'
lrwxrwxrwx 1 root root 0 Jun  2 16:17 uts -> 'uts:[4026532923]'
root@pop-os:/home/jaison# 


if you observer the uts is different this helps in identifing the namespaces.


The proc is also helpful for getting mounttable:
cat /proc/self/mountinfo