package resources

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func getDelegatedCgroupPath() (string, error) {
	data, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return "", err
	}
	uid := os.Getuid()
	userSlicePattern := fmt.Sprintf("/user.slice/user-%d.slice/user@%d.service", uid, uid)

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) >= 3 && parts[1] == "" { // cgroup v2
			cgroupPath := parts[2]
			if idx := strings.Index(cgroupPath, userSlicePattern); idx != -1 {
				resolved := filepath.Join("/sys/fs/cgroup", cgroupPath[:idx+len(userSlicePattern)])
				return resolved, nil
			}
		}
	}
	return "", fmt.Errorf("user slice cgroup not found in /proc/self/cgroup")
}

func isControllerAvailable(cgPath, controller string) bool {
	content, err := os.ReadFile(filepath.Join(cgPath, "cgroup.controllers"))
	if err != nil {
		return false
	}
	fields := strings.Fields(string(content))
	for _, f := range fields {
		if f == controller {
			return true
		}
	}
	return false
}

func enableSubtreeControllers(parent string) error {
	controllersBytes, err := os.ReadFile(filepath.Join(parent, "cgroup.controllers"))
	if err != nil {
		return err
	}
	controllers := strings.Fields(string(controllersBytes))
	var toEnable []string
	for _, ctrl := range controllers {
		if ctrl == "memory" || ctrl == "pids" || ctrl == "cpu" {
			toEnable = append(toEnable, "+"+ctrl)
		}
	}
	if len(toEnable) > 0 {
		subtreeControlPath := filepath.Join(parent, "cgroup.subtree_control")
		value := []byte(strings.Join(toEnable, " "))
		if err := os.WriteFile(subtreeControlPath, value, 0644); err != nil {
			return err
		}
	}
	return nil
}

func (cm *CgroupManager) ApplyLimits(limits *Limits) error {
	if limits == nil {
		return nil
	}

	useDefault := true
	if err := os.MkdirAll(cgroupRoot, 0755); err != nil {
		useDefault = false
	} else {
		testFile := filepath.Join(cgroupRoot, ".test")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			useDefault = false
		} else {
			_ = os.Remove(testFile)
		}
	}

	if !useDefault {
		delegated, err := getDelegatedCgroupPath()
		if err != nil {
			fmt.Println("[Warning] Resource limits skipped: cgroup writes not permitted in rootless mode.")
			cm.Path = ""
			return nil
		}

		if !isControllerAvailable(delegated, "memory") {
			fmt.Println("[Warning] Resource limits skipped: cgroup writes not permitted in rootless mode.")
			cm.Path = ""
			return nil
		}

		if err := enableSubtreeControllers(delegated); err != nil {
			fmt.Println("[Warning] Resource limits skipped: cgroup writes not permitted in rootless mode.")
			cm.Path = ""
			return nil
		}

		cageDir := filepath.Join(delegated, "cage")
		if err := os.MkdirAll(cageDir, 0755); err != nil {
			fmt.Println("[Warning] Resource limits skipped: cgroup writes not permitted in rootless mode.")
			cm.Path = ""
			return nil
		}

		if err := enableSubtreeControllers(cageDir); err != nil {
			fmt.Println("[Warning] Resource limits skipped: cgroup writes not permitted in rootless mode.")
			cm.Path = ""
			return nil
		}

		cm.Path = filepath.Join(cageDir, cm.Id)
	}

	if err := os.MkdirAll(cm.Path, 0755); err != nil {
		fmt.Println("[Warning] Resource limits skipped: cgroup writes not permitted in rootless mode.")
		cm.Path = ""
		return nil
	}

	if limits.MemoryMax > 0 {
		memFile := filepath.Join(cm.Path, "memory.max")
		if err := os.WriteFile(memFile, []byte(strconv.Itoa(int(limits.MemoryMax))), 0644); err != nil {
			if os.IsPermission(err) {
				fmt.Println("[Warning] Resource limits skipped: cgroup writes not permitted in rootless mode.")
				cm.Path = ""
				return nil
			}
		}
	}

	if limits.PidsMax > 0 {
		pidsFile := filepath.Join(cm.Path, "pids.max")
		if err := os.WriteFile(pidsFile, []byte(strconv.Itoa(limits.PidsMax)), 0644); err != nil {
			if os.IsPermission(err) {
				fmt.Println("[Warning] Resource limits skipped: cgroup writes not permitted in rootless mode.")
				cm.Path = ""
				return nil
			}
		}
	}

	if limits.CpuMax > 0 {
		cpuFile := filepath.Join(cm.Path, "cpu.max")
		quota := limits.CpuMax * 100000
		_ = os.WriteFile(cpuFile, []byte(fmt.Sprintf("%d 100000", quota)), 0644)
	}

	return nil
}

func (cm *CgroupManager) Destroy() error {
	if cm.Path == "" {
		return nil
	}
	return os.Remove(cm.Path)
}
