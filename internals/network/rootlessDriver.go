package network

import (
	"fmt"
	"os/exec"
	"syscall"

	"github.com/vishvananda/netlink"
)

type Slirp4netnsDriver struct {
	cmd *exec.Cmd
}

func (sd *Slirp4netnsDriver) ParentSetup(pid int, containerIP string, portMap *PortMapping) error {
	args := []string{
		("--device=tap0"),
		("-c"),
		fmt.Sprintf("%d", pid),
	}

	if portMap != nil {
		// Format: --port-forward=tcp:127.0.0.1:hostport:containerport
		args = append(args, fmt.Sprintf("--port-forward=tcp:0.0.0.0:%d:%d", portMap.HostPort, portMap.ContainerPort))
	}

	sd.cmd = exec.Command("slirp4netns", args...)

	if err := sd.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start slirp4netns daemon: %w", err)
	}
	return nil

}

func (sd *Slirp4netnsDriver) ChildSetup(containerIP string) error {
	// slirp4netns automatically configures the tap0 interface and routes
	// inside the container namespace once it attaches.
	// We just ensure the loopback interface is up.
	lo, err := netlink.LinkByName("lo")
	if err == nil {
		_ = netlink.LinkSetUp(lo)
	}
	return nil
}
func (sd *Slirp4netnsDriver) Teardown(pid int) error {
	if sd.cmd != nil && sd.cmd.Process != nil {
		// Stop the daemon process cleanly
		return sd.cmd.Process.Signal(syscall.SIGTERM)
	}
	return nil
}
