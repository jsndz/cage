package main

import (
	"cage/internals/cgroup"
	"cage/internals/network"
	"cage/internals/runtime"
	"cage/utils"
	"flag"
	"os"
)

func main() {
	id := utils.NewID()

	if len(os.Args) > 1 && os.Args[1] == "init" {
		runtime.InitContainer()
		return
	}
	cpu := flag.Int("cpu", 4, "maximum number of cpu cores")
	memory := flag.Int64("mem", 536870912, "maximum amount of memory needed")
	pids := flag.Int("pids", 100, "maximum number of pids")

	flag.Parse()

	limits := &cgroup.Limits{
		CpuMax:    *cpu,
		MemoryMax: *memory,
		PidsMax:   *pids,
	}

	bridge, err := network.GetorCreateBridge()
	if err != nil {
		panic(err)
	}

	runtime.StartContainer(id, limits, bridge)
}
