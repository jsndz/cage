package security

type SecurityConfig struct {
	Profile string

	CapAdd  []string
	CapDrop []string
}

func NewSecurityConfig(profile string, capAdd []string, capDrop []string) *SecurityConfig {
	return &SecurityConfig{
		Profile: profile,
		CapAdd:  capAdd,
		CapDrop: capDrop,
	}
}
