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

func NewContainerStatusBehavior(status ContainerStatus, c *LinuxContainer) (ContainerStatusBehavior, error) {
	switch status {
	case Created:
		return &CreatedState{c: c}, nil
	case Stopped:
		return &StoppedState{c: c}, nil
	case Running:
		return &RunningState{c: c}, nil
	default:
		return nil, fmt.Errorf("unknown status")
	}
}

func newStateTransitionError(from, to ContainerStatusBehavior) error {
	return &stateTransitionError{
		From: from.status().String(),
		To:   to.status().String(),
	}
}

// stateTransitionError is returned when an invalid state transition happens from one
// state to another.
type stateTransitionError struct {
	From string
	To   string
}

func (s *stateTransitionError) Error() string {
	return fmt.Sprintf("invalid state transition from %s to %s", s.From, s.To)
}

type ContainerStatusBehavior interface {
	transition(ContainerStatusBehavior) error
	destroy() error
	status() ContainerStatus
}

func destroy(c *LinuxContainer) error {
	logrus.Infof("destroying container...")
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
	c.statusBehavior = &StoppedState{c: c}
	logrus.Infof("destroy container complete")
	return err
}

// ******************************************************************************************
// 【StoppedState】 represents a container is a stopped/destroyed state.
// ******************************************************************************************
type StoppedState struct {
	c *LinuxContainer
}

func (b *StoppedState) status() ContainerStatus {
	return Stopped
}

func (b *StoppedState) transition(s ContainerStatusBehavior) error {
	switch s.(type) {
	case *RunningState:
		b.c.statusBehavior = s
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
	c *LinuxContainer
}

func (r *RunningState) status() ContainerStatus {
	return Running
}

func (r *RunningState) transition(s ContainerStatusBehavior) error {
	switch s.(type) {
	case *StoppedState:
		t, err := r.c.currentStatus()
		if err != nil {
			return err
		}
		if t == Running {
			return util.NewGenericError(fmt.Errorf("container still running"), util.ContainerNotStopped)
		}
		r.c.statusBehavior = s
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
	c *LinuxContainer
}

func (i *CreatedState) status() ContainerStatus {
	return Created
}

func (i *CreatedState) transition(s ContainerStatusBehavior) error {
	switch s.(type) {
	case *RunningState, *StoppedState:
		i.c.statusBehavior = s
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
