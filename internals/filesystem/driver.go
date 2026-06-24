package filesystem

type StorageDriver interface {
	Mount(lower, upper, merged, work string, readonly bool) error
	Clean(merged, overlayRoot string) error
}
