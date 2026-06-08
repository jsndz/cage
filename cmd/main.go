package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/unix"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		initContainer()
		return
	}

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

	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func initContainer() {
	lowerlayer := "/tmp/rootfs"
	upperlayer := "/tmp/overlay/upper"
	workdir := "/tmp/overlay/work"
	merged := "/tmp/overlay/merged"

	os.Mkdir(upperlayer, 0755)
	os.Mkdir(workdir, 0755)
	os.Mkdir(merged, 0755)

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
