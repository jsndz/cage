package cgroup

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type CgroupManager struct {
	Id   string
	Path string
}

const cgroupRoot = "/sys/fs/cgroup/cage"

func NewCgroupManager(id string) *CgroupManager {
	return &CgroupManager{
		Id:   id,
		Path: filepath.Join(cgroupRoot, id),
	}
}

func (cm *CgroupManager) ApplyLimits(limits *Limits) error {
	cg := "/sys/fs/cgroup/cage"
	os.MkdirAll(cg, 0755)
	os.WriteFile(cg+"/memory.max",
		[]byte(strconv.Itoa(int(limits.MemoryMax))),
		0644,
	)
	os.WriteFile(cg+"/pids.max",
		[]byte(strconv.Itoa(limits.PidsMax)),
		0644,
	)
	quota := limits.CpuMax * 100000
	os.WriteFile(cg+"/cpu.max", []byte(fmt.Sprintf("%d 100000", quota)), 0644)
	return nil
}

func (cm *CgroupManager) Destroy() error {
	return os.Remove(cm.Path)
}
