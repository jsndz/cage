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
	if err := network.SetUpContainerNetwork(pid, bridge, hostnet); err != nil {
		panic(err)
	}
	w.Write([]byte(containerIP))
	w.Close()
	cmd.Wait()

```


```go 

syncFile := os.NewFile(uintptr(3), "sync")
defer syncFile.Close()

ipBytes, err := io.ReadAll(syncFile)
if err != nil {
	panic(err)
}
containerIP := string(ipBytes)

network.SetUpVeth("eth0", containerIP)

```