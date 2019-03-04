package libcapsule

import (
	"github.com/sirupsen/logrus"
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

/**
创建一个ParentProcess实例，用于启动容器进程
有可能是InitParentProcess，也有可能是SetnsParentProcess
*/
func NewParentProcess(container *LinuxContainer, process *Process) (ParentProcess, error) {
	logrus.Infof("new parent process...")
	logrus.Infof("creating pipes...")
	// Config: parent 写，child(init process)读
	childConfigPipe, parentConfigPipe, err := os.Pipe()
	logrus.Infof("create config pipe complete, parentConfigPipe: %#v, configPipe: %#v", parentConfigPipe, childConfigPipe)

	cmd, err := container.buildCommand(process, childConfigPipe)
	logrus.Infof("build command complete, command: %#v", cmd)
	if err != nil {
		return nil, err
	}
	if process.Init {
		return NewParentInitProcess(process, cmd, parentConfigPipe, container), nil
	} else {
		return NewParentSetnsProcess(process, cmd, parentConfigPipe), nil
	}
}
