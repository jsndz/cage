package security

import (
	"cage/internals/filesystem"
)

func (s *SecurityConfig) ReadOnly(lowerdir, merged string) error {
	return filesystem.MountReadOnly(lowerdir, merged)
}
