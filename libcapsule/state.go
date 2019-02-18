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

func newStateTransitionError(from, to ContainerState) error {
	return &stateTransitionError{
		From: from.status().String(),
		To:   to.status().String(),
	}
}

// stateTransitionError is returned when an invalid containerState transition happens from one
// containerState to another.
type stateTransitionError struct {
	From string
	To   string
}

func (s *stateTransitionError) Error() string {
	return fmt.Sprintf("invalid containerState transition from %s to %s", s.From, s.To)
}

type ContainerState interface {
	transition(ContainerState) error
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
// 【StoppedState】 represents a container is a stopped/destroyed containerState.
// ******************************************************************************************
type StoppedState struct {
	c *LinuxContainerImpl
}

func (b *StoppedState) status() Status {
	return Stopped
}

func (b *StoppedState) transition(s ContainerState) error {
	switch s.(type) {
	case *RunningState:
		b.c.containerState = s
		return nil
	case *StoppedState:
		return nil
	}
	return newStateTransitionError(b, s)
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

func (r *RunningState) transition(s ContainerState) error {
	switch s.(type) {
	case *StoppedState:
		t, err := r.c.currentStatus()
		if err != nil {
			return err
		}
		if t == Running {
			return util.NewGenericError(fmt.Errorf("container still running"), util.ContainerNotStopped)
		}
		r.c.containerState = s
		return nil
	case *RunningState:
		return nil
	}
	return newStateTransitionError(r, s)
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

func (i *CreatedState) transition(s ContainerState) error {
	switch s.(type) {
	case *RunningState, *StoppedState:
		i.c.containerState = s
		return nil
	case *CreatedState:
		return nil
	}
	return newStateTransitionError(i, s)
}

func (i *CreatedState) destroy() error {
	i.c.initProcess.signal(unix.SIGKILL)
	return destroy(i.c)
}
