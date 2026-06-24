package filesystem

import (
	"fmt"
	"os"
	"os/exec"
)

type RootlessStorageDriver struct{}

func (d *RootlessStorageDriver) Mount(lower, upper, merged, work string, readonly bool) error {
	if readonly {
		// Bind mounts are permitted inside user-namespaces on modern kernels
		return MountReadOnly(lower, merged)
	}

	cmd := exec.Command("fuse-overlayfs",
		"-o", fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lower, upper, work),
		merged,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("fuse-overlayfs failed: %s (err: %w)", string(out), err)
	}
	return nil
}

func (d *RootlessStorageDriver) Clean(merged, overlayRoot string) error {
	cmd := exec.Command("fusermount", "-u", merged)
	if err := cmd.Run(); err != nil {
		// Fallback to standard unmount if fusermount fails
		return CleanOverlay(merged, overlayRoot)
	}
	return os.RemoveAll(overlayRoot)
}
