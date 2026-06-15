Pipe is a kernel provided channel for communication between processes.

Process A ----write----> [ PIPE ] ----read----> Process B

The kernel maintains a buffer for the pipe.


```go
	r, w, _ := os.Pipe()

	cmd.ExtraFiles = []*os.File{r}
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	pid := cmd.Process.Pid
	if err := os.WriteFile(
		"/sys/fs/cgroup/cage/cgroup.procs",
		[]byte(strconv.Itoa(pid)),
		0644,
	); err != nil {
		panic(err)
	}
	hostnet := "veth-host" + strconv.Itoa(pid)
	network.SetUpContainerNetwork(pid, bridge, "eth0", hostnet)
	w.Write([]byte{1})
	w.Close()
	cmd.Wait()

```


```go 

syncFile := os.NewFile(uintptr(3), "sync")

buf := make([]byte, 1)

syncFile.Read(buf)
network.SetUpVeth("eth0")

```