package security

import (
	"fmt"
	"syscall"

	libseccomp "github.com/seccomp/libseccomp-golang"
)

var defaultBlockedSyscalls = []string{
	// Process accounting
	"acct",

	// Kernel keyring
	"add_key",
	"keyctl",
	"request_key",

	// eBPF programs
	"bpf",

	// System clock manipulation
	"clock_adjtime",
	"clock_settime",
	"settimeofday",
	"stime",

	// New clone API (harder to filter args)
	"clone3",

	// Kernel module management
	"create_module",
	"delete_module",
	"finit_module",
	"init_module",
	"get_kernel_syms",
	"query_module",

	// NUMA memory policy
	"get_mempolicy",
	"mbind",
	"set_mempolicy",

	// I/O port permissions
	"ioperm",
	"iopl",

	// Kernel object comparison
	"kcmp",

	// Load new kernel
	"kexec_file_load",
	"kexec_load",

	// Kernel profiling
	"lookup_dcookie",

	// Mount manipulation
	"mount_setattr",
	"move_mount",
	"open_tree",
	"pivot_root",
	"umount",
	"umount2",

	// NFS server control
	"nfsservctl",

	// Perf events
	"perf_event_open",

	// Disk quota
	"quotactl",
	"quotactl_fd",

	// Reboot
	"reboot",

	// Namespace manipulation
	"setns",
	"unshare",

	// Swap management
	"swapon",
	"swapoff",

	// Deprecated/legacy syscalls
	"sysfs",
	"_sysctl",
	"uselib",
	"ustat",

	// Userspace page fault handling
	"userfaultfd",

	// Virtual 8086 mode (x86 only)
	"vm86",
	"vm86old",
}

var sandboxAllowedSyscalls = []string{
	// File I/O
	"read",
	"write",
	"open",
	"openat",
	"close",
	"stat",
	"fstat",
	"lstat",
	"newfstatat",
	"lseek",
	"access",
	"faccessat",
	"faccessat2",
	"readlink",
	"readlinkat",
	"getcwd",
	"readv",
	"writev",
	"pread64",
	"pwrite64",

	// Memory management
	"mmap",
	"mprotect",
	"munmap",
	"brk",
	"mremap",
	"madvise",

	// Process identity
	"exit",
	"exit_group",
	"getpid",
	"getppid",
	"getuid",
	"geteuid",
	"getgid",
	"getegid",
	"gettid",
	"set_tid_address",
	"set_robust_list",
	"get_robust_list",

	// Signals
	"rt_sigaction",
	"rt_sigprocmask",
	"rt_sigreturn",
	"sigaltstack",
	"kill",

	// Directory operations
	"getdents",
	"getdents64",
	"mkdir",
	"mkdirat",
	"rmdir",
	"rename",
	"renameat",
	"renameat2",
	"unlink",
	"unlinkat",
	"link",
	"linkat",
	"symlink",
	"symlinkat",

	// File metadata operations
	"chmod",
	"fchmod",
	"fchmodat",
	"chown",
	"fchown",
	"fchownat",
	"utimensat",
	"truncate",
	"ftruncate",
	"fallocate",

	// Pipe and file descriptor operations
	"pipe",
	"pipe2",
	"dup",
	"dup2",
	"dup3",
	"fcntl",
	"ioctl",

	// Network
	"socket",
	"connect",
	"accept",
	"accept4",
	"bind",
	"listen",
	"sendto",
	"recvfrom",
	"sendmsg",
	"recvmsg",
	"shutdown",
	"setsockopt",
	"getsockopt",
	"getsockname",
	"getpeername",
	"select",
	"pselect6",
	"poll",
	"ppoll",
	"epoll_create",
	"epoll_create1",
	"epoll_ctl",
	"epoll_wait",
	"epoll_pwait",
	"eventfd",
	"eventfd2",

	// Time (read-only)
	"clock_gettime",
	"clock_getres",
	"gettimeofday",
	"nanosleep",
	"clock_nanosleep",

	// Exec and wait
	"execve",
	"execveat",
	"wait4",
	"waitid",

	// Misc essentials
	"arch_prctl",
	"prctl",
	"futex",
	"clone",
	"fork",
	"vfork",
	"uname",
	"sysinfo",
	"getrandom",
	"rseq",
	"seccomp",
	"prlimit64",
	"setrlimit",
	"getrlimit",
	"umask",
	"chdir",
	"fchdir",
	"statfs",
	"fstatfs",
	"statx",
	"close_range",
	"memfd_create",
	"copy_file_range",
	"sendfile",
	"splice",
	"tee",
}

// Profiles:
//   - "privileged": No seccomp filtering is applied. The process retains full
//     syscall access.
//   - "default" (or empty/unknown): Allows all syscalls except a curated blocklist
//     of dangerous operations (kernel module loading, namespace manipulation,
//     reboot, etc.). Blocked syscalls return EPERM.
//   - "sandbox": Denies all syscalls by default (returning EPERM) and only permits
//     a minimal whitelist of syscalls needed for typical containerized workloads.
func (s *SecurityConfig) SetUpSeccomp() error {
	switch s.Profile {
	case "privileged":
		// No seccomp filtering — the container has unrestricted syscall access.
		return nil

	case "sandbox":
		return s.setupSandboxProfile()

	default:
		// Covers "default", empty string, and any unknown profile name.
		return s.setupDefaultProfile()
	}
}

func (s *SecurityConfig) setupDefaultProfile() error {
	filter, err := libseccomp.NewFilter(libseccomp.ActAllow)
	if err != nil {
		return fmt.Errorf("seccomp: failed to create default filter: %w", err)
	}
	defer filter.Release()

	blockAction := libseccomp.ActErrno.SetReturnCode(int16(syscall.EPERM))

	for _, name := range defaultBlockedSyscalls {
		call, err := libseccomp.GetSyscallFromName(name)
		if err != nil {
			// Syscall does not exist on this architecture (e.g. vm86 on arm64).
			continue
		}
		if err := filter.AddRule(call, blockAction); err != nil {
			return fmt.Errorf("seccomp: failed to add block rule for %q: %w", name, err)
		}
	}

	if err := filter.Load(); err != nil {
		return fmt.Errorf("seccomp: failed to load default filter: %w", err)
	}

	return nil
}

func (s *SecurityConfig) setupSandboxProfile() error {
	defaultAction := libseccomp.ActErrno.SetReturnCode(int16(syscall.EPERM))

	filter, err := libseccomp.NewFilter(defaultAction)
	if err != nil {
		return fmt.Errorf("seccomp: failed to create sandbox filter: %w", err)
	}
	defer filter.Release()

	for _, name := range sandboxAllowedSyscalls {
		call, err := libseccomp.GetSyscallFromName(name)
		if err != nil {
			// Syscall does not exist on this architecture.
			continue
		}
		if err := filter.AddRule(call, libseccomp.ActAllow); err != nil {
			return fmt.Errorf("seccomp: failed to add allow rule for %q: %w", name, err)
		}
	}

	if err := filter.Load(); err != nil {
		return fmt.Errorf("seccomp: failed to load sandbox filter: %w", err)
	}

	return nil
}
