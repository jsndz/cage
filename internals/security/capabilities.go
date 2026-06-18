package security

import "github.com/syndtr/gocapability/capability"

var CapabilityMap = make(map[string]capability.Cap)

func init() {
	for c := capability.Cap(0); c <= capability.CAP_CHECKPOINT_RESTORE; c++ {
		CapabilityMap[c.String()] = c
	}
}

func (s *SecurityConfig) SetUpCapabilities(pid int) {
	caps, _ := capability.NewPid2(pid)
	for _, cap := range s.CapAdd {
		caps.Set(capability.PERMITTED, CapabilityMap[cap])
		caps.Set(capability.EFFECTIVE, CapabilityMap[cap])
	}
	for _, cap := range s.CapDrop {
		caps.Unset(capability.PERMITTED, CapabilityMap[cap])
		caps.Unset(capability.EFFECTIVE, CapabilityMap[cap])
	}
	caps.Apply(capability.CAPS)
}
