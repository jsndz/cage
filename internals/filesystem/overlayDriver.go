package filesystem

type OverlayDriver struct {
}

func (d *OverlayDriver) Mount(lower, upper, merged, work string, readonly bool) error {
	if readonly {
		return MountReadOnly(lower, merged)
	}
	return MountOverlay(lower, upper, work, merged)
}

func (d *OverlayDriver) Clean(merged, overlayRoot string) error {
	return CleanOverlay(merged, overlayRoot)
}
