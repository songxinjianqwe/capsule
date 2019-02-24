package libcapsule

import (
	"fmt"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"github.com/songxinjianqwe/rune/libcapsule/util/system"
	"golang.org/x/sys/unix"
	"os"
	"syscall"
	"time"
)

func NewNoChildProcessWrapper(initProcessPid int, initProcessStartTime uint64, c *LinuxContainerImpl) ProcessWrapper {
	return &NoChildProcessWrapperImpl{
		initProcessPid:       initProcessPid,
		initProcessStartTime: initProcessStartTime,
		container:            c,
	}
}

/**
ProcessWrapper接口的实现类，没有行为
*/
type NoChildProcessWrapperImpl struct {
	initProcessPid       int
	initProcessStartTime uint64
	container            *LinuxContainerImpl
}

func (p *NoChildProcessWrapperImpl) pid() int {
	return p.initProcessPid
}

func (p *NoChildProcessWrapperImpl) start() error {
	panic("implement me")
}

func (p *NoChildProcessWrapperImpl) terminate() error {
	panic("implement me")
}

func (p *NoChildProcessWrapperImpl) wait() error {
	// https://stackoverflow.com/questions/1157700/how-to-wait-for-exit-of-non-children-processes
	// 无法使用wait之类的系统调用来等待一个无关进程的结束
	// 可以轮询 /prod/${pid}
	for {
		<-time.After(time.Millisecond * 100)
		stat, err := system.GetProcessStat(p.pid())
		// 如果出现err，或者进程已经成为僵尸进程，则退出循环
		if os.IsNotExist(err) {
			return nil
		}
		if err != nil || stat.State == system.Zombie {
			return err
		}
	}
}

func (p *NoChildProcessWrapperImpl) startTime() (uint64, error) {
	return p.initProcessStartTime, nil
}

func (p *NoChildProcessWrapperImpl) signal(sig os.Signal) error {
	s, ok := sig.(syscall.Signal)
	if !ok {
		return util.NewGenericError(fmt.Errorf("os: unsupported signal type:%v", sig), util.SystemError)
	}
	return unix.Kill(p.pid(), s)
}
