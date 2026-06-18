package runtime

import (
	"time"
)

type SandBox struct {
	ID        string
	Pid       int
	Status    string
	IpAddr    string
	Command   []string
	CreatedAt time.Time
	Cgroup    string
	Rootfs    string
}

type Manager struct {
	sandboxes map[string]*SandBox
}

func NewManager() *Manager {
	return &Manager{
		sandboxes: make(map[string]*SandBox),
	}
}

func CreateSandbox(id string) *SandBox {

	sandbox := &SandBox{
		ID:        id,
		Status:    "creating",
		CreatedAt: time.Now(),
	}
	return sandbox
}
