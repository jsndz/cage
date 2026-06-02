package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	cmd := exec.Command("sh", "-c", "echo $$; sleep 5")

	hostPIDNS, _ := os.Readlink("/proc/self/ns/pid")
	hostMNTNS, _ := os.Readlink("/proc/self/ns/mnt")
	hostUTSNS, _ := os.Readlink("/proc/self/ns/uts")
	flags :=
		syscall.CLONE_NEWPID | // PID namespace
			syscall.CLONE_NEWNS | // Mount namespace
			syscall.CLONE_NEWUTS // Hostname (UTS) namespace
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: uintptr(flags),
	}

	// check for isolation

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	childPID := cmd.Process.Pid

	pidNS, _ := os.Readlink(fmt.Sprintf("/proc/%d/ns/pid", childPID))
	mntNS, _ := os.Readlink(fmt.Sprintf("/proc/%d/ns/mnt", childPID))
	utsNS, _ := os.Readlink(fmt.Sprintf("/proc/%d/ns/uts", childPID))

	fmt.Println("Child PID NS:", pidNS)
	fmt.Println("Child MNT NS:", mntNS)
	fmt.Println("Child UTS NS:", utsNS)

	data, _ := io.ReadAll(stdout)

	cmd.Wait()

	fmt.Println("Host PID NS:", hostPIDNS)
	fmt.Println("Host MNT NS:", hostMNTNS)
	fmt.Println("Host UTS NS:", hostUTSNS)

	fmt.Print(string(data))
}
