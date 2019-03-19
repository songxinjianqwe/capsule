package libcapsule

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
)

/*
对于Exec来说，start返回后，非daemon的进程已经结束了。
*/
func execStartHook(p *ParentAbstractProcess) error {
	// 如果启用了cgroups，那么将exec进程的pid也加进去
	if len(p.container.cgroupManager.GetPaths()) > 0 {
		if err := p.container.cgroupManager.JoinCgroupSet(p.pid()); err != nil {
			return exception.NewGenericErrorWithContext(err, exception.CgroupsError, fmt.Sprintf("adding pid %d to cgroups", p.pid()))
		}
	}

	// exec process会在启动后阻塞，直至收到config
	if err := p.sendConfigAndClosePipe(); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.PipeError, "sending config to init process")
	}

	// 如果是detach，则直接结束。
	if !p.detach() {
		logrus.Infof("wait child process exit...")
		if err := p.wait(); err != nil {
			return exception.NewGenericErrorWithContext(err, exception.CmdWaitError, "waiting child process exit")
		}
		logrus.Infof("child process exited")
	}
	return nil
}
