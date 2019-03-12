package libcapsule

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/network"
	"os"
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

func destroy(c *LinuxContainer) (err error) {
	logrus.Infof("destroying container...")
	logrus.Infof("destroying cgroup manager...")
	err = c.cgroupManager.Destroy()
	if err != nil {
		logrus.Warnf("destroy cgroup manager failed, cause: %s", err.Error())
	}
	logrus.Infof("destroying endpoint...")
	if c.endpoint != nil {
		if err := network.Disconnect(c.endpoint); err != nil {
			logrus.Warnf("destroy cgroup manager failed, cause: %s", err.Error())
		}
	}
	logrus.Infof("removing container root files...")
	removeErr := os.RemoveAll(c.root)
	if removeErr != nil {
		logrus.Warnf("remove container root runtime files failed, cause: %s", err.Error())
		err = removeErr
	}
	c.parentProcess = nil
	c.statusBehavior = &StoppedStatusBehavior{c: c}
	logrus.Infof("destroy container complete")
	return err
}
