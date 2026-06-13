package filesystem

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// MountOverlay mounts the overlay filesystem.
func MountOverlay(lower, upper, work, merged string) error {
	if err := os.MkdirAll(upper, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(work, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(merged, 0755); err != nil {
		return err
	}

	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lower, upper, work)
	return unix.Mount("overlay", merged, "overlay", 0, opts)
}

// CleanOverlay unmounts the merged directory and removes the overlay directories.
func CleanOverlay(merged, overlayRoot string) error {
	if err := unix.Unmount(merged, 0); err != nil {
		return err
	}
	return os.RemoveAll(overlayRoot)
}

// PivotRoot isolates the container's root filesystem.
func PivotRoot(rootfs string) error {
	if err := unix.Mount(
		"",
		"/",
		"",
		unix.MS_PRIVATE|unix.MS_REC,
		"",
	); err != nil {
		return err
	}

	if err := unix.Mount(
		rootfs,
		rootfs,
		"",
		unix.MS_BIND|unix.MS_REC,
		"",
	); err != nil {
		return err
	}

	putOld := filepath.Join(rootfs, ".oldroot")
	if err := os.MkdirAll(putOld, 0755); err != nil {
		return err
	}

	if err := unix.PivotRoot(rootfs, putOld); err != nil {
		return err
	}

	if err := os.Chdir("/"); err != nil {
		return err
	}

	if err := SetupSystemMounts(); err != nil {
		return err
	}

	if err := unix.Unmount("/.oldroot", unix.MNT_DETACH); err != nil {
		return err
	}

	return os.RemoveAll("/.oldroot")
}

// SetupSystemMounts mounts proc, sysfs, and tmpfs inside the container's root.
func SetupSystemMounts() error {
	if err := os.MkdirAll("/proc", 0555); err != nil {
		return err
	}
	if err := os.MkdirAll("/sys", 0555); err != nil {
		return err
	}
	if err := os.MkdirAll("/dev", 0755); err != nil {
		return err
	}

	if err := unix.Mount(
		"proc",
		"/proc",
		"proc",
		0,
		"",
	); err != nil {
		return err
	}

	if err := unix.Mount(
		"sysfs",
		"/sys",
		"sysfs",
		unix.MS_NOSUID|unix.MS_NOEXEC|unix.MS_NODEV,
		"",
	); err != nil {
		return err
	}

	return unix.Mount(
		"tmpfs",
		"/dev",
		"tmpfs",
		0,
		"",
	)
}
