package libcapsule

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
