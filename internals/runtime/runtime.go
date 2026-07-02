package runtime

import (
	"cage/internals/filesystem"
	"cage/internals/network"
	"cage/internals/resources"
	"cage/internals/security"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"golang.org/x/sys/unix"
)

type initPayload struct {
	ContainerIP     string                   `json:"container_ip"`
	SecurityConfig  *security.SecurityConfig `json:"security_config"`
	ApparmorProfile string                   `json:"apparmor_profile"`
}

const lowerlayer = "/tmp/rootfs"
const upperlayer = "/tmp/overlay/upper"
const workdir = "/tmp/overlay/work"
const merged = "/tmp/overlay/merged"

// StartContainer sets up cgroups, clones namespaces, setup network and runs the container.
func StartContainer(containerID string, limits *resources.Limits, portMap *network.PortMapping, securityConfig *security.SecurityConfig) {
	sb := CreateSandbox(containerID)
	cm := resources.NewCgroupManager(containerID)
	if err := cm.ApplyLimits(limits); err != nil {
		panic(err)
	}
	sb.Cgroup = cm.Path
	containerIP, err := network.FindFreeIP()
	if err != nil {
		panic(err)
	}
	var networkDriver network.NetworkDriver
	var storageDriver filesystem.StorageDriver
	if securityConfig.Rootless {
		networkDriver = &network.Slirp4netnsDriver{}
		storageDriver = &filesystem.RootlessStorageDriver{}
	} else {
		networkDriver = &network.BridgedDriver{}
		storageDriver = &filesystem.OverlayDriver{}
	}
	// Load AppArmor profile into the kernel (must happen in parent before child exec's)
	apparmorProfile, err := securityConfig.LoadApparmorProfile(containerID)
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("/proc/self/exe", "init")

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if securityConfig.Rootless {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags: uintptr(
				syscall.CLONE_NEWPID |
					syscall.CLONE_NEWNS |
					syscall.CLONE_NEWUTS |
					syscall.CLONE_NEWNET |
					syscall.CLONE_NEWUSER,
			),
			UidMappings: []syscall.SysProcIDMap{
				{
					ContainerID: 0,
					HostID:      os.Getuid(),
					Size:        1,
				},
			},

			GidMappings: []syscall.SysProcIDMap{
				{
					ContainerID: 0,
					HostID:      os.Getgid(),
					Size:        1,
				},
			},

			GidMappingsEnableSetgroups: false,
		}

	} else {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags: uintptr(
				syscall.CLONE_NEWPID |
					syscall.CLONE_NEWNS |
					syscall.CLONE_NEWUTS |
					syscall.CLONE_NEWNET,
			),
		}
	}
	r, w, _ := os.Pipe()

	cmd.ExtraFiles = []*os.File{r}
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	pid := cmd.Process.Pid
	sb.Pid = pid
	sb.Status = "starting"
	fmt.Print("Container PID: ", pid, "\n")

	if cm.Path != "" {
		if err := os.WriteFile(
			filepath.Join(cm.Path, "cgroup.procs"),
			[]byte(strconv.Itoa(pid)),
			0644,
		); err != nil {
			panic(err)
		}
	}

	if err := networkDriver.ParentSetup(pid, containerIP, portMap); err != nil {
		panic(err)
	}
	// Send config (IP + security) to child via pipe as JSON
	payload := initPayload{
		ContainerIP:     containerIP,
		SecurityConfig:  securityConfig,
		ApparmorProfile: apparmorProfile,
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		panic(err)
	}

	sb.IpAddr = containerIP
	sb.Status = "running"

	w.Close()
	cmd.Wait()

	if err := cm.Destroy(); err != nil {
		panic(err)
	}
	if err := networkDriver.TearDown(pid); err != nil {
		panic(err)
	}
	if err := storageDriver.Clean(merged, filepath.Dir(merged)); err != nil {
		panic(err)
	}
	// Unload AppArmor profile from the kernel
	if err := securityConfig.UnloadApparmorProfile(containerID); err != nil {
		panic(err)
	}
}

// InitContainer initializes the isolated environment inside namespaces.
// It receives its configuration (IP, security) from the parent via a pipe.
func InitContainer() {

	// Read configuration from parent via pipe
	syncFile := os.NewFile(uintptr(3), "sync")
	defer syncFile.Close()

	var payload initPayload
	if err := json.NewDecoder(syncFile).Decode(&payload); err != nil {
		panic(err)
	}
	var networkDriver network.NetworkDriver
	var storageDriver filesystem.StorageDriver
	if payload.SecurityConfig.Rootless {
		networkDriver = &network.Slirp4netnsDriver{}
		storageDriver = &filesystem.RootlessStorageDriver{}
	} else {
		networkDriver = &network.BridgedDriver{}
		storageDriver = &filesystem.OverlayDriver{}
	}
	if err := storageDriver.Mount(lowerlayer, upperlayer, merged, workdir, payload.SecurityConfig.Readonly); err != nil {
		panic(err)
	}
	if !payload.SecurityConfig.Readonly {
		if err := os.MkdirAll(merged+"/etc", 0755); err != nil {
			panic(err)
		}
		if err := os.WriteFile(merged+"/etc/resolv.conf", []byte("nameserver 8.8.8.8\n"), 0644); err != nil {
			panic(err)
		}
	}
	if payload.SecurityConfig.Rootless {
		if err := filesystem.ChrootRoot(merged); err != nil {
			panic(err)
		}
		if err := filesystem.SetupSystemMounts(); err != nil {
			panic(err)
		}
	} else {
		if err := filesystem.PivotRoot(merged); err != nil {
			panic(err)
		}
	}
	if payload.SecurityConfig.Readonly {
		if err := unix.Mount("", "/", "", unix.MS_REMOUNT|unix.MS_BIND|unix.MS_RDONLY, ""); err != nil {
			panic(err)
		}
	}
	if err := unix.Sethostname([]byte("cage")); err != nil {
		panic(err)
	}

	if err := networkDriver.ChildSetup(payload.ContainerIP); err != nil {
		panic(err)
	}

	// Set up capabilities inside the child process (pid=0 means current process)
	if err := payload.SecurityConfig.SetUpCapabilities(); err != nil {
		panic(err)
	}

	// Install seccomp-bpf syscall filter based on the security profile
	if err := payload.SecurityConfig.SetUpSeccomp(); err != nil {
		panic(err)
	}

	// Apply AppArmor profile — transitions on next exec()
	if err := payload.SecurityConfig.ApplyApparmorProfile(payload.ApparmorProfile); err != nil {
		panic(err)
	}

	// Prevent the process from gaining new privileges after exec
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		panic(err)
	}

	if err := syscall.Exec(
		"/bin/sh",
		[]string{"/bin/sh"},
		os.Environ(),
	); err != nil {
		panic(err)
	}
}
