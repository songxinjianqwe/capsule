package libcapsule

import (
	"fmt"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
)

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
		t, err := behavior.c.detectContainerStatus()
		if err != nil {
			return err
		}
		if t == Running {
			return exception.NewGenericError(fmt.Errorf("container still running"), exception.ContainerStillRunningError)
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
		return exception.NewGenericError(fmt.Errorf("container is still running, cant be destroyed"), exception.ContainerStillRunningError)
	}
	return destroy(behavior.c)
}
