package libcapsule

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
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

/**
创建一个ParentProcess实例，用于启动容器进程
有可能是InitParentProcess，也有可能是ExecParentProcess
*/
func NewParentProcess(container *LinuxContainer, process *Process) (ParentProcess, error) {
	logrus.Infof("new parent process...")
	logrus.Infof("creating pipes...")
	// Config: parent 写，child(init process)读
	childConfigPipe, parentConfigPipe, err := os.Pipe()
	logrus.Infof("create config pipe complete, parentConfigPipe: %#v, configPipe: %#v", parentConfigPipe, childConfigPipe)

	cmd, err := container.buildCommand(process, childConfigPipe)
	if err != nil {
		return nil, err
	}
	if process.Init {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", EnvInitializerType, string(StandardInitializer)))
		logrus.Infof("build command complete, command: %#v", cmd)
		logrus.Infof("new parent init process...")
		return &ParentInitProcess{
			initProcessCmd:   cmd,
			parentConfigPipe: parentConfigPipe,
			container:        container,
			process:          process,
		}, nil
	} else {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", EnvInitializerType, string(ExecInitializer)))
		logrus.Infof("build command complete, command: %#v", cmd)
		logrus.Infof("new parent exec process...")
		return &ParentExecProcess{
			execProcessCmd:   cmd,
			parentConfigPipe: parentConfigPipe,
			container:        container,
			process:          process,
		}, nil
	}
}

// **************************************************************************************************
// util
// **************************************************************************************************

func sendConfig(containerConfig configs.ContainerConfig, process Process, id string, pipe *os.File) error {
	initConfig := &InitConfig{
		ContainerConfig: containerConfig,
		ProcessConfig:   process,
		ID:              id,
	}
	logrus.Infof("sending config: %#v", initConfig)
	bytes, err := json.Marshal(initConfig)
	if err != nil {
		return err
	}
	_, err = pipe.WriteString(string(bytes))
	return err
}
