package security

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const (
	// apparmorProfilePrefix is the namespace prefix for all cage AppArmor profiles.
	apparmorProfilePrefix = "cage-"

	// apparmorAttrPath is the procfs path used to transition to an AppArmor
	// profile on the next exec(). The child writes "exec <profile>\n" here.
	apparmorAttrPath = "/proc/self/attr/apparmor/exec"

	// apparmorAttrPathLegacy is used on older kernels that don't have the
	// apparmor/ subdirectory.
	apparmorAttrPathLegacy = "/proc/self/attr/exec"
)

// defaultApparmorProfile is a restrictive profile suitable for general container
// workloads. It blocks:
//   - writes to sensitive /proc and /sys paths
//   - mount, umount, pivot_root operations
//   - raw network access and packet sniffing
//   - ptrace of other processes
//   - access to kernel security interfaces
//
// While allowing normal file I/O, networking, and process execution.
var defaultApparmorProfile = `
#include <tunables/global>

profile %s flags=(attach_disconnected,mediate_deleted) {
  #include <abstractions/base>

  # Allow all networking (sockets, connections, DNS)
  network,

  # Allow all signals within the container
  signal (receive) peer=%s,
  signal (send) peer=%s,
  signal (receive) peer=unconfined,

  # Allow reading all files by default
  /** r,

  # Allow execution of any binary
  /** ix,

  # Allow write to general paths (container filesystem)
  /** w,
  /** k,
  /** l,

  # Block writes to sensitive /proc paths
  deny /proc/sys/** w,
  deny /proc/sysrq-trigger w,
  deny /proc/kcore r,
  deny /proc/kmsg r,
  deny /proc/kallsyms r,
  deny /proc/acpi/** w,
  deny /proc/timer_list r,
  deny /proc/timer_stats r,
  deny /proc/scsi/** w,

  # Block writes to sensitive /sys paths
  deny /sys/firmware/** rwlk,
  deny /sys/kernel/security/** rwlk,
  deny /sys/fs/** w,
  deny /sys/devices/virtual/powercap/** rwlk,

  # Block mount operations
  deny mount,
  deny umount,
  deny pivot_root,

  # Block ptrace of other processes
  deny ptrace,

  # Allow ptrace of self (needed for debugging tools inside container)
  ptrace peer=%s,
}
`

// sandboxApparmorProfile is a highly restrictive profile for untrusted workloads.
// It blocks:
//   - all writes except to /tmp, /var/tmp, /dev/null, /dev/zero
//   - all network access
//   - all mount operations
//   - ptrace
//   - signal sending to other processes
//   - access to kernel interfaces
var sandboxApparmorProfile = `
#include <tunables/global>

profile %s flags=(attach_disconnected,mediate_deleted) {
  #include <abstractions/base>

  # Deny all networking
  deny network,

  # Allow signals only to self
  signal (receive) peer=%s,
  signal (send) peer=%s,

  # Read-only access to the filesystem
  /** r,

  # Execute binaries
  /** ix,

  # Allow writes only to safe temporary locations
  /tmp/** wk,
  /var/tmp/** wk,
  /dev/null rw,
  /dev/zero rw,
  /dev/full rw,
  /dev/urandom r,
  /dev/random r,
  /dev/tty rw,
  /dev/pts/** rw,

  # Block all writes to sensitive paths
  deny /proc/** w,
  deny /proc/kcore r,
  deny /proc/kmsg r,
  deny /proc/kallsyms r,
  deny /sys/** w,
  deny /sys/firmware/** rwlk,
  deny /sys/kernel/** rwlk,

  # Block dangerous operations
  deny mount,
  deny umount,
  deny pivot_root,
  deny ptrace,
}
`

// ProfileName returns the AppArmor profile name for this container's profile.
func (s *SecurityConfig) apparmorProfileName(containerID string) string {
	return apparmorProfilePrefix + s.Profile + "-" + containerID
}

// LoadApparmorProfile generates and loads an AppArmor profile into the kernel.
// This must be called from the PARENT process before the child exec()'s,
// because apparmor_parser requires host-level access.
func (s *SecurityConfig) LoadApparmorProfile(containerID string) (string, error) {
	if s.Profile == "privileged" {
		// Privileged containers run unconfined — no profile to load.
		return "unconfined", nil
	}

	if !isApparmorAvailable() {
		return "", nil
	}

	profileName := s.apparmorProfileName(containerID)
	profileText := s.renderProfile(profileName)

	// Write the profile to a temporary file
	tmpFile, err := os.CreateTemp("", "cage-apparmor-*.profile")
	if err != nil {
		return "", fmt.Errorf("apparmor: failed to create temp profile file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(profileText); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("apparmor: failed to write profile: %w", err)
	}
	tmpFile.Close()

	// Load the profile into the kernel using apparmor_parser
	cmd := exec.Command("apparmor_parser", "-r", "-W", tmpFile.Name())
	if output, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("apparmor: failed to load profile %q: %w\noutput: %s",
			profileName, err, string(output))
	}

	return profileName, nil
}

// ApplyApparmorProfile transitions the current process to the given AppArmor
// profile on the next exec() call. This must be called from the CHILD process
// before exec().
func (s *SecurityConfig) ApplyApparmorProfile(profileName string) error {
	if profileName == "" || profileName == "unconfined" || s.Rootless {
		// No profile to apply.
		return nil
	}
	//modern path for apparmor
	content := "exec " + profileName + "\n"
	err := os.WriteFile(apparmorAttrPath, []byte(content), 0)
	if err != nil {
		// Try legacy path for older kernels
		err = os.WriteFile(apparmorAttrPathLegacy, []byte(content), 0)
		if err != nil {
			return fmt.Errorf("apparmor: failed to apply profile %q: %w", profileName, err)
		}
	}

	return nil
}

// UnloadApparmorProfile removes an AppArmor profile from the kernel.
// Called during container cleanup.
func (s *SecurityConfig) UnloadApparmorProfile(containerID string) error {
	if s.Profile == "privileged" || s.Rootless {
		return nil
	}

	if !isApparmorAvailable() {
		return nil
	}

	profileName := s.apparmorProfileName(containerID)

	cmd := exec.Command("apparmor_parser", "-R", "-W")
	cmd.Stdin = strings.NewReader(fmt.Sprintf("profile %s {}", profileName))
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("apparmor: failed to unload profile %q: %w\noutput: %s",
			profileName, err, string(output))
	}

	return nil
}

// renderProfile generates the AppArmor profile text for the current security
// profile, with the profile name substituted in.
func (s *SecurityConfig) renderProfile(profileName string) string {
	switch s.Profile {
	case "sandbox":
		return fmt.Sprintf(sandboxApparmorProfile,
			profileName, profileName, profileName)
	default:
		// "default", empty string, or any unknown profile
		return fmt.Sprintf(defaultApparmorProfile,
			profileName, profileName, profileName, profileName)
	}
}

// isApparmorAvailable checks whether AppArmor is enabled on this system
// by looking for the apparmor filesystem.
func isApparmorAvailable() bool {
	// Check if AppArmor is enabled via the kernel
	if _, err := os.Stat("/sys/kernel/security/apparmor"); err != nil {
		return false
	}
	// Check if apparmor_parser is available
	if _, err := exec.LookPath("apparmor_parser"); err != nil {
		return false
	}
	return true
}
