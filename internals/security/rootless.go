package security

func (s *SecurityConfig) RunRootless() bool {
	return s.Rootless
}
