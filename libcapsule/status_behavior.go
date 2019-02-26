package libcapsule

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"golang.org/x/sys/unix"
	"os"
	"syscall"
	"time"
)

func NewContainerStatusBehavior(status ContainerStatus, c *LinuxContainer) (ContainerStatusBehavior, error) {
	switch status {
	case Created:
		return &CreatedStatusBehavior{c: c}, nil
	case Stopped:
		return &StoppedStatusBehavior{c: c}, nil
	case Running:
		return &RunningStatusBehavior{c: c}, nil
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
	logrus.Infof("destroying cgroup manager...")
	err := c.cgroupManager.Destroy()
	logrus.Infof("removing container root files...")
	if rerr := os.RemoveAll(c.root); err == nil {
		err = rerr
	}
	c.initProcess = nil
	c.statusBehavior = &StoppedStatusBehavior{c: c}
	logrus.Infof("destroy container complete")
	return err
}

// ******************************************************************************************
// 【StoppedStatusBehavior】 represents a container is a stopped/destroyed state.
// ******************************************************************************************
type StoppedStatusBehavior struct {
	c *LinuxContainer
}

func (behavior *StoppedStatusBehavior) status() ContainerStatus {
	return Stopped
}

func (behavior *StoppedStatusBehavior) transition(s ContainerStatusBehavior) error {
	switch s.(type) {
	case *RunningStatusBehavior:
		behavior.c.statusBehavior = s
		return nil
	case *StoppedStatusBehavior:
		return nil
	}
	return newStateTransitionError(behavior, s)
}

func (behavior *StoppedStatusBehavior) destroy() error {
	return destroy(behavior.c)
}

// ******************************************************************************************
// 【RunningStatusBehavior】 represents a container that is currently running.
// ******************************************************************************************
type RunningStatusBehavior struct {
	c *LinuxContainer
}

func (behavior *RunningStatusBehavior) status() ContainerStatus {
	return Running
}

func (behavior *RunningStatusBehavior) transition(s ContainerStatusBehavior) error {
	switch s.(type) {
	case *StoppedStatusBehavior:
		t, err := behavior.c.currentStatus()
		if err != nil {
			return err
		}
		if t == Running {
			return util.NewGenericError(fmt.Errorf("container still running"), util.ContainerNotStopped)
		}
		behavior.c.statusBehavior = s
		return nil
	case *RunningStatusBehavior:
		return nil
	}
	return newStateTransitionError(behavior, s)
}

func (behavior *RunningStatusBehavior) destroy() error {
	t, err := behavior.c.currentStatus()
	if err != nil {
		return err
	}
	if t == Running {
		return util.NewGenericError(fmt.Errorf("container is not destroyed"), util.ContainerNotStopped)
	}
	return destroy(behavior.c)
}

// ******************************************************************************************
// 【CreatedStatusBehavior】
// ******************************************************************************************
type CreatedStatusBehavior struct {
	c *LinuxContainer
}

func (behavior *CreatedStatusBehavior) status() ContainerStatus {
	return Created
}

func (behavior *CreatedStatusBehavior) transition(s ContainerStatusBehavior) error {
	switch s.(type) {
	case *RunningStatusBehavior, *StoppedStatusBehavior:
		behavior.c.statusBehavior = s
		return nil
	case *CreatedStatusBehavior:
		return nil
	}
	return newStateTransitionError(behavior, s)
}

func (behavior *CreatedStatusBehavior) destroy() error {
	logrus.Infof("send SIGKILL signal to init process")
	_ = behavior.c.Signal(unix.SIGKILL)
	// 最多等10s
	for i := 0; i < 100; i++ {
		logrus.Infof("[%d]detect container status", i)
		time.Sleep(100 * time.Millisecond)
		// 检测容器状态，如果容器被杀掉了，那么出现err，此时就可以destroy容器了（一种检测容器状态的方法）
		if err := behavior.c.Signal(syscall.Signal(0)); err != nil {
			logrus.Infof("container was killed, destroying...")
			return destroy(behavior.c)
		}
	}
	return fmt.Errorf("container init still running")
}
