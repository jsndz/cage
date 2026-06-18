package runtime

import (
	"cage/internals/filesystem"
	"cage/internals/network"
	"cage/internals/resources"
	"cage/internals/security"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// StartContainer sets up cgroups, clones namespaces, setup network and runs the container.
func StartContainer(containerID string, limits *resources.Limits, bridge *netlink.Bridge, portMap *network.PortMapping, securityConfig *security.SecurityConfig) {
	sb := CreateSandbox(containerID)
	cm := resources.NewCgroupManager(containerID)
	if err := cm.ApplyLimits(limits); err != nil {
		panic(err)
	}
	sb.Cgroup = cm.Path
	containerIP, err := network.FindFreeIP()
	if err != nil {
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
	sb.Pid = pid
	sb.Status = "starting"
	fmt.Print("Container PID: ", pid, "\n")

	if err := os.WriteFile(
		filepath.Join(cm.Path, "cgroup.procs"),
		[]byte(strconv.Itoa(pid)),
		0644,
	); err != nil {
		panic(err)
	}
	securityConfig.SetUpCapabilities(pid)
	hostnet := "veth-h" + strconv.Itoa(pid)
	if err := network.SetUpContainerNetwork(pid, bridge, hostnet); err != nil {
		panic(err)
	}
	if _, err := w.Write([]byte(containerIP)); err != nil {
		panic(err)
	}

	sb.IpAddr = containerIP
	sb.Status = "running"
	fmt.Println(portMap)
	if portMap != nil {
		ip := containerIP
		if idx := strings.Index(ip, "/"); idx != -1 {
			ip = ip[:idx]
		}
		if err := network.AddPortForwarding(portMap.HostPort, ip, portMap.ContainerPort, bridge.Attrs().Name); err != nil {
			panic(err)
		}
	}

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

	// Write default resolv.conf to configure DNS resolver (e.g., 8.8.8.8) inside the container
	if err := os.MkdirAll(merged+"/etc", 0755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(merged+"/etc/resolv.conf", []byte("nameserver 8.8.8.8\n"), 0644); err != nil {
		panic(err)
	}

	if err := filesystem.PivotRoot(merged); err != nil {
		panic(err)
	}
	syncFile := os.NewFile(uintptr(3), "sync")
	defer syncFile.Close()

	ipBytes, err := io.ReadAll(syncFile)
	if err != nil {
		panic(err)
	}
	containerIP := string(ipBytes)

	if err := network.SetUpVeth("eth0", containerIP); err != nil {
		panic(err)
	}

	if err := syscall.Exec(
		"/bin/sh",
		[]string{"/bin/sh"},
		os.Environ(),
	); err != nil {
		panic(err)
	}
}
