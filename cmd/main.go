package main

import (
	"cage/internals/cgroup"
	"cage/internals/runtime"
	"flag"
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

	flag.Parse()

	limits := cgroup.Limits{
		CpuMax:    *cpu,
		MemoryMax: *memory,
		PidsMax:   *pids,
	}

	runtime.StartContainer(&limits)
}
