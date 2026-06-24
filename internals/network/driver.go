package network

type PortMapping struct {
	HostPort      int
	ContainerPort int
}

type Driver interface {
	ParentSetup(pid int, containerIP string, portMap *PortMapping) error
	ChildSetup(containerIP string) error
	TearDown(pid int) error
}
