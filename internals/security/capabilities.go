package security

import (
	"fmt"
	"strings"

	"github.com/syndtr/gocapability/capability"
)

var CapabilityMap = make(map[string]capability.Cap)

func init() {
	for c := capability.Cap(0); c <= capability.CAP_CHECKPOINT_RESTORE; c++ {
		CapabilityMap["CAP_"+strings.ToUpper(c.String())] = c
	}
}

func (s *SecurityConfig) SetUpCapabilities() error {
	caps, err := capability.NewPid2(0)
	if err != nil {
		return err
	}

	caps.Clear(capability.CAPS)
	caps.Clear(capability.BOUNDS)
	var cps []string
	switch s.Profile {
	case "default":
		cps = DefaultCaps
	case "sandbox":
		cps = SandboxCaps
	case "privileged":
		cps = PrivilegedCaps
	default:
		cps = DefaultCaps
	}
	for _, cap := range cps {
		if _, ok := CapabilityMap[cap]; !ok {
			return fmt.Errorf("invalid capability: %s", cap)
		}
		caps.Set(capability.BOUNDS, CapabilityMap[cap])
		caps.Set(capability.PERMITTED, CapabilityMap[cap])
		caps.Set(capability.EFFECTIVE, CapabilityMap[cap])
	}

	for _, cap := range s.CapAdd {
		if _, ok := CapabilityMap[cap]; !ok {
			return fmt.Errorf("invalid capability: %s", cap)
		}
		caps.Set(capability.BOUNDS, CapabilityMap[cap])
		caps.Set(capability.PERMITTED, CapabilityMap[cap])
		caps.Set(capability.EFFECTIVE, CapabilityMap[cap])
	}
	for _, cap := range s.CapDrop {
		if _, ok := CapabilityMap[cap]; !ok {
			return fmt.Errorf("invalid capability: %s", cap)
		}
		caps.Unset(capability.BOUNDS, CapabilityMap[cap])
		caps.Unset(capability.PERMITTED, CapabilityMap[cap])
		caps.Unset(capability.EFFECTIVE, CapabilityMap[cap])
	}
	return caps.Apply(capability.CAPS | capability.BOUNDS)

}
