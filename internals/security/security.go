package security

type SecurityConfig struct {
	Profile  string
	Readonly bool
	Rootless bool
	CapAdd   []string
	CapDrop  []string
}

var DefaultCaps = []string{
	"CAP_CHOWN",
	"CAP_DAC_OVERRIDE",
	"CAP_FOWNER",
	"CAP_FSETID",

	"CAP_KILL",

	"CAP_SETGID",
	"CAP_SETUID",
	"CAP_SETPCAP",

	"CAP_NET_BIND_SERVICE",

	"CAP_SETFCAP",

	"CAP_AUDIT_WRITE",

	"CAP_MKNOD",
}

var SandboxCaps = []string{
	// empty: no capabilities
}

var PrivilegedCaps = []string{
	"CAP_CHOWN",
	"CAP_DAC_OVERRIDE",
	"CAP_DAC_READ_SEARCH",
	"CAP_FOWNER",
	"CAP_FSETID",
	"CAP_KILL",

	"CAP_SETGID",
	"CAP_SETUID",
	"CAP_SETPCAP",

	"CAP_LINUX_IMMUTABLE",

	"CAP_NET_BIND_SERVICE",
	"CAP_NET_BROADCAST",
	"CAP_NET_ADMIN",
	"CAP_NET_RAW",

	"CAP_IPC_LOCK",
	"CAP_IPC_OWNER",

	"CAP_SYS_MODULE",
	"CAP_SYS_RAWIO",
	"CAP_SYS_CHROOT",
	"CAP_SYS_PTRACE",
	"CAP_SYS_PACCT",
	"CAP_SYS_ADMIN",
	"CAP_SYS_BOOT",
	"CAP_SYS_NICE",
	"CAP_SYS_RESOURCE",
	"CAP_SYS_TIME",
	"CAP_SYS_TTY_CONFIG",

	"CAP_MKNOD",
	"CAP_LEASE",

	"CAP_AUDIT_WRITE",
	"CAP_AUDIT_CONTROL",
	"CAP_AUDIT_READ",

	"CAP_SETFCAP",

	"CAP_MAC_OVERRIDE",
	"CAP_MAC_ADMIN",

	"CAP_SYSLOG",
	"CAP_WAKE_ALARM",
	"CAP_BLOCK_SUSPEND",

	"CAP_PERFMON",
	"CAP_BPF",
	"CAP_CHECKPOINT_RESTORE",
}

func NewSecurityConfig(profile string, capAdd []string, capDrop []string, readonly bool) *SecurityConfig {

	return &SecurityConfig{
		Profile:  profile,
		Readonly: readonly,
		CapAdd:   capAdd,
		CapDrop:  capDrop,
	}
}
