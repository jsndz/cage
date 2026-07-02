package main

import (
	"cage/internals/network"
	"cage/internals/resources"
	"cage/internals/runtime"
	"cage/internals/security"
	"cage/utils"
	"flag"
	"fmt"
	"os"
	"strings"
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
	profile := flag.String("profile", "", "linux capablities profile(e.g. default, restricted, privileged)")
	addCap := flag.String("cap-add", "", "add linux capabilities(e.g. CAP_SYS_ADMIN,CAP_NET_ADMIN)")
	dropCap := flag.String("cap-drop", "", "drop linux capabilites(e.g. )")
	readonly := flag.Bool("read-only", false, "mount rootfs as read-only")
	rootless := flag.Bool("rootless", false, "run container in rootless mode")
	flag.Parse()

	limits := &resources.Limits{
		CpuMax:    *cpu,
		MemoryMax: *memory,
		PidsMax:   *pids,
	}
	securityConfig := &security.SecurityConfig{
		Profile:  *profile,
		Readonly: *readonly,
		Rootless: *rootless,
		CapAdd:   []string{},
		CapDrop:  []string{},
	}
	if *addCap != "" {
		securityConfig.CapAdd = append(securityConfig.CapAdd, strings.Split(*addCap, ",")...)
	}
	if *dropCap != "" {
		securityConfig.CapDrop = append(securityConfig.CapDrop, strings.Split(*dropCap, ",")...)
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
	id := utils.NewID()
	runtime.StartContainer(id, limits, portMap, securityConfig)
}
