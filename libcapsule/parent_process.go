package libcapsule

import (
	"encoding/json"
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
