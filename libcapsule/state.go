package libcapsule

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/config"
	"github.com/songxinjianqwe/rune/libcapsule/util"
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

func destroy(c *LinuxContainer) error {
	if !c.config.Namespaces.Contains(config.NEWPID) {
		if err := signalAllProcesses(c.cgroupManager, unix.SIGKILL); err != nil {
			logrus.Warn(err)
		}
	}
	err := c.cgroupManager.Destroy()
	if rerr := os.RemoveAll(c.root); err == nil {
		err = rerr
	}
	c.initProcess = nil
	c.containerState = &stoppedState{c: c}
	return err
}

// stoppedState represents a container is a stopped/destroyed containerState.
type stoppedState struct {
	c *LinuxContainer
}

func (b *stoppedState) status() Status {
	return Stopped
}

func (b *stoppedState) transition(s ContainerState) error {
	switch s.(type) {
	case *runningState:
		b.c.containerState = s
		return nil
	case *stoppedState:
		return nil
	}
	return newStateTransitionError(b, s)
}

func (b *stoppedState) destroy() error {
	return destroy(b.c)
}

// runningState represents a container that is currently running.
type runningState struct {
	c *LinuxContainer
}

func (r *runningState) status() Status {
	return Running
}

func (r *runningState) transition(s ContainerState) error {
	switch s.(type) {
	case *stoppedState:
		t, err := r.c.currentStatus()
		if err != nil {
			return err
		}
		if t == Running {
			return util.NewGenericError(fmt.Errorf("container still running"), util.ContainerNotStopped)
		}
		r.c.containerState = s
		return nil
	case *runningState:
		return nil
	}
	return newStateTransitionError(r, s)
}

func (r *runningState) destroy() error {
	t, err := r.c.currentStatus()
	if err != nil {
		return err
	}
	if t == Running {
		return util.NewGenericError(fmt.Errorf("container is not destroyed"), util.ContainerNotStopped)
	}
	return destroy(r.c)
}

type createdState struct {
	c *LinuxContainer
}

func (i *createdState) status() Status {
	return Created
}

func (i *createdState) transition(s ContainerState) error {
	switch s.(type) {
	case *runningState, *stoppedState:
		i.c.containerState = s
		return nil
	case *createdState:
		return nil
	}
	return newStateTransitionError(i, s)
}

func (i *createdState) destroy() error {
	i.c.initProcess.signal(unix.SIGKILL)
	return destroy(i.c)
}
