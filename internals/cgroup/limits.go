package cgroup

type Limits struct {
	MemoryMax int64
	CpuMax    int
	PidsMax   int
}
