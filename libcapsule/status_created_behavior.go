package libcapsule

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"syscall"
	"time"
)

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
		logrus.Infof("[try %d times]detect container status", i)
		time.Sleep(100 * time.Millisecond)
		// 检测容器状态，如果容器被杀掉了，那么出现err，此时就可以destroy容器了（一种检测容器状态的方法）
		if err := behavior.c.Signal(syscall.Signal(0)); err != nil {
			logrus.Infof("container was killed, destroying...")
			return destroy(behavior.c)
		}
	}
	return fmt.Errorf("container init still running")
}
