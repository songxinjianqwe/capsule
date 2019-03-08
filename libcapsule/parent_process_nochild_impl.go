package libcapsule

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/util/proc"
	"os"
	"time"
)

func NewParentNoChildProcess(initProcessPid int, initProcessStartTime uint64, c *LinuxContainer) ParentProcess {
	return &ParentNoChildProcess{
		initProcessPid:       initProcessPid,
		initProcessStartTime: initProcessStartTime,
		container:            c,
	}
}

/*
ParentProcess接口的实现类，no child意味着我们现在启动的进程并不是容器init process的父进程，但仍是init process的代理
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

/*
不需要实现
*/
func (p *ParentNoChildProcess) start() error {
	return errors.New("should not be called")
}

/*
不需要实现
*/
func (p *ParentNoChildProcess) terminate() error {
	return errors.New("should not be called")
}

func (p *ParentNoChildProcess) wait() error {
	// https://stackoverflow.com/questions/1157700/how-to-wait-for-exit-of-non-children-processes
	// 无法使用wait之类的系统调用来等待一个无关进程的结束
	// 可以轮询 /prod/${pid}/stat
	// 如果该文件不存在，则说明进程已停止
	for {
		<-time.After(time.Millisecond * 100)
		stat, err := proc.GetProcessStat(p.pid())
		// 如果出现err，或者进程已经成为僵尸进程，则退出循环
		if os.IsNotExist(err) {
			logrus.Infof("%d process exited(/proc/%d/stat not exists)", p.pid(), p.pid())
			return nil
		}
		if err != nil || stat.Status == proc.Zombie {
			return err
		}
	}
}

func (p *ParentNoChildProcess) startTime() (uint64, error) {
	return p.initProcessStartTime, nil
}

func (p *ParentNoChildProcess) signal(sig os.Signal) error {
	process, err := os.FindProcess(p.pid())
	if err != nil {
		return err
	}
	logrus.Infof("send %s to %d", sig, p.pid())
	return process.Signal(sig)
}
