package filesystem

import (
	"errors"
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

// MountReadOnly bind-mounts the rootfs as read-only (no overlay upper layer)
// and configures the resolv.conf file safely.
func MountReadOnly(lower, merged string) error {
	if _, err := os.Stat(lower); err != nil {
		return errors.New("lower layer does not exist")
	}

	// Create host-side resolv.conf in the parent of merged (overlayRoot)
	overlayDir := filepath.Dir(merged)
	if err := os.MkdirAll(overlayDir, 0755); err != nil {
		return err
	}
	resolvFile := filepath.Join(overlayDir, "resolv.conf")
	if err := os.WriteFile(resolvFile, []byte("nameserver 8.8.8.8\n"), 0644); err != nil {
		return err
	}

	if err := os.MkdirAll(merged, 0755); err != nil {
		return err
	}

	// Bind mount the lower dir
	// mounting lower to merged
	if err := unix.Mount(lower, merged, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
		return err
	}

	// Ensure merged/etc/resolv.conf exists so we can mount over it
	resolvConfPath := filepath.Join(merged, "etc/resolv.conf")
	if err := os.MkdirAll(filepath.Dir(resolvConfPath), 0755); err != nil {
		return err
	}
	if _, err := os.Stat(resolvConfPath); os.IsNotExist(err) {
		f, err := os.Create(resolvConfPath)
		if err != nil {
			return err
		}
		f.Close()
	}

	// Bind-mount the resolv.conf file on top of the placeholder
	if err := unix.Mount(resolvFile, resolvConfPath, "", unix.MS_BIND, ""); err != nil {
		return err
	}

	// Remount as read-only
	return unix.Mount("", merged, "", unix.MS_REMOUNT|unix.MS_BIND|unix.MS_RDONLY, "")
}

// CleanOverlay unmounts the merged directory and removes the overlay directories.
func CleanOverlay(merged, overlayRoot string) error {
	if err := unix.Unmount(merged, unix.MNT_DETACH); err != nil {
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
