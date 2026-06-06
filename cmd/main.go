package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/unix"
)

func main() {
	args := os.Args
	cmd := exec.Command(args[1], args[2:]...)

	flags :=
		syscall.CLONE_NEWPID | // PID namespace
			syscall.CLONE_NEWNS | // Mount namespace
			syscall.CLONE_NEWUTS // Hostname (UTS) namespace

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: uintptr(flags),
	}

}

func InitContainer() {

	rootfs := "/tmp/rootfs"
	if err := unix.Mount("", "/", "", unix.MS_PRIVATE|unix.MS_REC, ""); err != nil {
		panic(err)
	}
	// change the dir to mount point
	// pivot_root takes mount point and mount point is entry in a mount table
	// binding to itself to make it a mount point
	if err := unix.Mount(rootfs, rootfs, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
		panic(err)
	}
	// use pivot for switching root
	putOld := filepath.Join(rootfs, ".oldroot")
	if err := os.MkdirAll(putOld, 0755); err != nil {
		panic(err)
	}
	if err := unix.PivotRoot(rootfs, putOld); err != nil {
		panic(err)
	}
	// since you are creating a file sys you need
	// proc, sys, dev and bin for binary and lib

	CreateFilesystem() // for now or can use base image

	// container is centered around a process
	// process has a root pointer to  /
	// so it like a pointer having swtiched around the / and /rootfs
	// switched place but oldroot is pointing to /
	// unmount that
	if err := unix.Unmount(putOld, unix.MNT_DETACH); err != nil {
		panic(err)
	}

}

func CreateFilesystem() {
	// create proc
	err := os.Mkdir("/proc", 0555)
	if err != nil {
		panic(err)
	}
	if err := unix.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		panic(err)
	}
	err = os.Mkdir("/sys", 0555)
	if err != nil {
		panic(err)
	}
	if err := unix.Mount(
		"sysfs",
		"/sys",
		"sysfs",
		uintptr(unix.MS_NOSUID|unix.MS_NOEXEC|unix.MS_NODEV),
		"",
	); err != nil {
		panic(err)
	}
	err = os.Mkdir("/proc", 0555)
	if err != nil {
		panic(err)
	}
	if err := unix.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		panic(err)
	}
}
