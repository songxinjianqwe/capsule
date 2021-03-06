package libcapsule

import (
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
	"os"
)

type ParentProcess interface {
	// pid returns the pid for the running process.
	pid() int

	// create starts the process execution.
	start() error

	// send a SIGKILL to the process and wait for the exit.
	terminate() error

	// wait waits on the process returning the process state.
	wait() error

	// startTime returns the process create time.
	startTime() (uint64, error)

	// send signal to the process
	signal(os.Signal) error

	// detach returns the process is detach
	detach() bool
}

type InitExecConfig struct {
	ContainerConfig configs.ContainerConfig `json:"container_config"`
	ProcessConfig   Process                 `json:"process_config"`
	ID              string                  `json:"id"`
}
