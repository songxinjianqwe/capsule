package libcapsule

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"github.com/songxinjianqwe/rune/libcapsule/util/system"
	"golang.org/x/sys/unix"
	"os"
	"syscall"
	"time"
)

func NewNoChildProcessWrapper(initProcessPid int, initProcessStartTime uint64, c *LinuxContainer) ParentProcess {
	return &ParentNoChildProcess{
		initProcessPid:       initProcessPid,
		initProcessStartTime: initProcessStartTime,
		container:            c,
	}
}

/**
ProcessWrapper接口的实现类，no child意味着我们现在启动的进程并不是容器init process的父进程，但仍是init process的代理
*/
type ParentNoChildProcess struct {
	initProcessPid       int
	initProcessStartTime uint64
	container            *LinuxContainer
}

func (p *ParentNoChildProcess) detach() bool {
	return false
}

func (p *ParentNoChildProcess) pid() int {
	return p.initProcessPid
}

func (p *ParentNoChildProcess) start() error {
	panic("implement me")
}

func (p *ParentNoChildProcess) terminate() error {
	panic("implement me")
}

func (p *ParentNoChildProcess) wait() error {
	// https://stackoverflow.com/questions/1157700/how-to-wait-for-exit-of-non-children-processes
	// 无法使用wait之类的系统调用来等待一个无关进程的结束
	// 可以轮询 /prod/${pid}/stat
	// 如果该文件不存在，则说明进程已停止
	for {
		<-time.After(time.Millisecond * 100)
		stat, err := system.GetProcessStat(p.pid())
		// 如果出现err，或者进程已经成为僵尸进程，则退出循环
		if os.IsNotExist(err) {
			logrus.Infof("%d process exited(/proc/%d/stat not exists)", p.pid(), p.pid())
			return nil
		}
		if err != nil || stat.State == system.Zombie {
			return err
		}
	}
}

func (p *ParentNoChildProcess) startTime() (uint64, error) {
	return p.initProcessStartTime, nil
}

func (p *ParentNoChildProcess) signal(sig os.Signal) error {
	s, ok := sig.(syscall.Signal)
	if !ok {
		return util.NewGenericError(fmt.Errorf("os: unsupported signal type:%v", sig), util.SystemError)
	}
	return unix.Kill(p.pid(), s)
}
