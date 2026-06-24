package network

import (
	"strconv"
	"strings"
)

type BridgedDriver struct {
}

func (bd *BridgedDriver) ParentSetup(pid int, containerIP string, portMap *PortMapping) error {
	bridge, err := GetorCreateBridge()
	if err != nil {
		return err
	}
	hostnet := "veth-h" + strconv.Itoa(pid)
	if err := SetUpContainerNetwork(pid, bridge, hostnet); err != nil {
		panic(err)
	}
	if portMap != nil {
		ip := containerIP
		if idx := strings.Index(ip, "/"); idx != -1 {
			ip = ip[:idx]
		}
		if err := AddPortForwarding(portMap.HostPort, ip, portMap.ContainerPort, bridge.Attrs().Name); err != nil {
			panic(err)
		}
	}
	return nil
}
func (bd *BridgedDriver) ChildSetup(containerIP string) error {
	err := SetUpVeth("eth0", containerIP)
	return err
}

func (bd *BridgedDriver) TearDown(pid int) error {
	hostnet := "veth-h" + strconv.Itoa(pid)
	if err := CleanBridge(hostnet); err != nil {
		return err
	}
	return nil
}
