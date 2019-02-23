package libcapsule

import (
	"fmt"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"golang.org/x/sys/unix"
	"os"
	"syscall"
)

func NewNoChildProcessWrapper(initProcessPid int, initProcessStartTime uint64) ProcessWrapper {
	return &NoChildProcessWrapperImpl{
		initProcessPid:       initProcessPid,
		initProcessStartTime: initProcessStartTime,
	}
}

/**
ProcessWrapper接口的实现类，没有行为
*/
type NoChildProcessWrapperImpl struct {
	initProcessPid       int
	initProcessStartTime uint64
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

func (p *NoChildProcessWrapperImpl) wait() (*os.ProcessState, error) {
	panic("implement me")
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
