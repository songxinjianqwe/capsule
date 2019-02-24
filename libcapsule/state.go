package libcapsule

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"github.com/songxinjianqwe/rune/libcapsule/util/system"
	"golang.org/x/sys/unix"
	"os"
)

func NewContainerState(statusStr string, c *LinuxContainerImpl) (ContainerState, error) {
	status := statusFromString(statusStr)
	switch status {
	case Created:
		return &CreatedState{c: c}, nil
	case Stopped:
		return &StoppedState{c: c}, nil
	case Running:
		return &RunningState{c: c}, nil
	default:
		return nil, fmt.Errorf("Unknown status")
	}
}

type ContainerState interface {
	destroy() error
	status() Status
}

func destroy(c *LinuxContainerImpl) error {
	if !c.config.Namespaces.Contains(configc.NEWPID) {
		if err := system.SignalAllProcesses(c.cgroupManager, unix.SIGKILL); err != nil {
			logrus.Warn(err)
		}
	}
	err := c.cgroupManager.Destroy()
	if rerr := os.RemoveAll(c.root); err == nil {
		err = rerr
	}
	c.initProcess = nil
	c.containerState = &StoppedState{c: c}
	return err
}

// ******************************************************************************************
// 【StoppedState】 represents a container is a stopped/destroyed state.
// ******************************************************************************************
type StoppedState struct {
	c *LinuxContainerImpl
}

func (b *StoppedState) status() Status {
	return Stopped
}

func (b *StoppedState) destroy() error {
	return destroy(b.c)
}

// ******************************************************************************************
// 【RunningState】 represents a container that is currently running.
// ******************************************************************************************
type RunningState struct {
	c *LinuxContainerImpl
}

func (r *RunningState) status() Status {
	return Running
}

func (r *RunningState) destroy() error {
	t, err := r.c.currentStatus()
	if err != nil {
		return err
	}
	if t == Running {
		return util.NewGenericError(fmt.Errorf("container is not destroyed"), util.ContainerNotStopped)
	}
	return destroy(r.c)
}

// ******************************************************************************************
// 【CreatedState】
// ******************************************************************************************
type CreatedState struct {
	c *LinuxContainerImpl
}

func (i *CreatedState) status() Status {
	return Created
}

func (i *CreatedState) destroy() error {
	i.c.initProcess.signal(unix.SIGKILL)
	return destroy(i.c)
}
