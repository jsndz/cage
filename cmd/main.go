package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"golang.org/x/sys/unix"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		initContainer()
		return
	}
	cpu := flag.Int("cpu", 4, "maximum number of cpu cores")
	memory := flag.Int("mem", 536870912, "maximum amount of memory needed")
	pids := flag.Int("pids", 100, "maximum number of pids")

	flag.Parse()
	cg := "/sys/fs/cgroup/cage"
	defer os.RemoveAll(cg)
	os.MkdirAll(cg, 0755)
	os.WriteFile(cg+"/memory.max",
		[]byte(strconv.Itoa(*memory)),
		0644,
	)
	os.WriteFile(cg+"/pids.max",
		[]byte(strconv.Itoa(*pids)),
		0644,
	)
	quota := *cpu * 100000
	os.WriteFile(cg+"/cpu.max", []byte(fmt.Sprintf("%d 100000", quota)), 0644)

	cmd := exec.Command("/proc/self/exe", "init")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: uintptr(
			syscall.CLONE_NEWPID |
				syscall.CLONE_NEWNS |
				syscall.CLONE_NEWUTS,
		),
	}

	if err := cmd.Start(); err != nil {
		panic(err)
	}
	pid := cmd.Process.Pid
	os.WriteFile(
		"/sys/fs/cgroup/cage/cgroup.procs",
		[]byte(strconv.Itoa(pid)),
		0644,
	)
	cmd.Wait()
	unix.Unmount("/tmp/overlay/merged", 0)
	os.RemoveAll("/tmp/overlay")
	os.Remove(cg)
}

func initContainer() {
	lowerlayer := "/tmp/rootfs"
	upperlayer := "/tmp/overlay/upper"
	workdir := "/tmp/overlay/work"
	merged := "/tmp/overlay/merged"

	os.MkdirAll(upperlayer, 0755)
	os.MkdirAll(workdir, 0755)
	os.MkdirAll(merged, 0755)

	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerlayer, upperlayer, workdir)
	if err := unix.Mount("overlay", merged, "overlay", 0, opts); err != nil {
		panic(err)
	}

	rootfs := merged
	if err := unix.Sethostname([]byte("cage")); err != nil {
		panic(err)
	}

	if err := unix.Mount(
		"",
		"/",
		"",
		unix.MS_PRIVATE|unix.MS_REC,
		"",
	); err != nil {
		panic(err)
	}

	if err := unix.Mount(
		rootfs,
		rootfs,
		"",
		unix.MS_BIND|unix.MS_REC,
		"",
	); err != nil {
		panic(err)
	}

	putOld := filepath.Join(rootfs, ".oldroot")

	if err := os.MkdirAll(putOld, 0755); err != nil {
		panic(err)
	}

	if err := unix.PivotRoot(rootfs, putOld); err != nil {
		panic(err)
	}

	if err := os.Chdir("/"); err != nil {
		panic(err)
	}

	setupFilesystem()

	if err := unix.Unmount("/.oldroot", unix.MNT_DETACH); err != nil {
		panic(err)
	}

	if err := os.RemoveAll("/.oldroot"); err != nil {
		panic(err)
	}

	if err := syscall.Exec(
		"/bin/bash",
		[]string{"/bin/bash"},
		os.Environ(),
	); err != nil {
		panic(err)
	}
}

func setupFilesystem() {
	os.MkdirAll("/proc", 0555)
	os.MkdirAll("/sys", 0555)
	os.MkdirAll("/dev", 0755)

	if err := unix.Mount(
		"proc",
		"/proc",
		"proc",
		0,
		"",
	); err != nil {
		panic(err)
	}

	if err := unix.Mount(
		"sysfs",
		"/sys",
		"sysfs",
		unix.MS_NOSUID|unix.MS_NOEXEC|unix.MS_NODEV,
		"",
	); err != nil {
		panic(err)
	}

	if err := unix.Mount(
		"tmpfs",
		"/dev",
		"tmpfs",
		0,
		"",
	); err != nil {
		panic(err)
	}
}
