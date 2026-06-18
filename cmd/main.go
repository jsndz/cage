package main

import (
	"cage/internals/cgroup"
	"cage/internals/network"
	"cage/internals/runtime"
	"cage/utils"
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		runtime.InitContainer()
		return
	}
	cpu := flag.Int("cpu", 4, "maximum number of cpu cores")
	memory := flag.Int64("mem", 536870912, "maximum amount of memory needed")
	pids := flag.Int("pids", 100, "maximum number of pids")
	portMapStr := flag.String("p", "", "port forwarding mapping (e.g. 8080:80)")

	flag.Parse()

	limits := &cgroup.Limits{
		CpuMax:    *cpu,
		MemoryMax: *memory,
		PidsMax:   *pids,
	}

	var portMap *network.PortMapping
	if *portMapStr != "" {
		var hostPort, containerPort int
		_, err := fmt.Sscanf(*portMapStr, "%d:%d", &hostPort, &containerPort)
		if err != nil {
			panic("Invalid port mapping format. Expected hostPort:containerPort (e.g., 8080:80)")
		}
		portMap = &network.PortMapping{
			HostPort:      hostPort,
			ContainerPort: containerPort,
		}
	}

	bridge, err := network.GetorCreateBridge()
	if err != nil {
		panic(err)
	}
	id := utils.NewID()
	runtime.StartContainer(id, limits, bridge, portMap)
}
