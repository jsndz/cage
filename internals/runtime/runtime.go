package runtime

import (
	"cage/internals/cgroup"
	"cage/internals/filesystem"
	"cage/internals/network"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// StartContainer sets up cgroups, clones namespaces, setup network and runs the container.
func StartContainer(limits *cgroup.Limits, bridge *netlink.Bridge) {
	cm := cgroup.NewCgroupManager("cage1")
	if err := cm.ApplyLimits(limits); err != nil {
		panic(err)
	}

	cmd := exec.Command("/proc/self/exe", "init")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: uintptr(
			syscall.CLONE_NEWPID |
				syscall.CLONE_NEWNS |
				syscall.CLONE_NEWUTS |
				syscall.CLONE_NEWNET,
		),
	}
	r, w, _ := os.Pipe()

	cmd.ExtraFiles = []*os.File{r}
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	pid := cmd.Process.Pid
	fmt.Print("Container PID: ", pid, "\n")
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

	if err := filesystem.CleanOverlay("/tmp/overlay/merged", "/tmp/overlay"); err != nil {
		panic(err)
	}

	if err := cm.Destroy(); err != nil {
		panic(err)
	}
	if err := network.CleanBridge(hostnet); err != nil {
		panic(err)
	}
}

// InitContainer initializes the isolated environment inside namespaces.
func InitContainer() {
	lowerlayer := "/tmp/rootfs"
	upperlayer := "/tmp/overlay/upper"
	workdir := "/tmp/overlay/work"
	merged := "/tmp/overlay/merged"

	if err := filesystem.MountOverlay(lowerlayer, upperlayer, workdir, merged); err != nil {
		panic(err)
	}

	if err := unix.Sethostname([]byte("cage")); err != nil {
		panic(err)
	}

	if err := filesystem.PivotRoot(merged); err != nil {
		panic(err)
	}
	syncFile := os.NewFile(uintptr(3), "sync")

	buf := make([]byte, 1)

	syncFile.Read(buf)
	network.SetUpVeth("eth0")

	if err := syscall.Exec(
		"/bin/bash",
		[]string{"/bin/bash"},
		os.Environ(),
	); err != nil {
		panic(err)
	}
}
