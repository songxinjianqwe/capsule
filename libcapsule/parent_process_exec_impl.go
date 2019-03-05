package libcapsule

import (
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/util"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
	"github.com/songxinjianqwe/capsule/libcapsule/util/proc"
	"os"
	"os/exec"
	"strings"
)

/**
ParentProcess接口的实现类，包裹了ExecProcess
*/
type ParentExecProcess struct {
	execProcessCmd   *exec.Cmd
	parentConfigPipe *os.File
	container        *LinuxContainer
	process          *Process
}

/**
对于Exec来说，start返回后，非daemon的进程已经结束了。
*/
func (p *ParentExecProcess) start() error {
	logrus.Infof("ParentExecProcess starting...")
	err := p.execProcessCmd.Start()
	if err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "starting init process command")
	}
	logrus.Infof("exec process started, EXEC_PROCESS_PID: [%d]", p.pid())
	// exec process会在启动后阻塞，直至收到namespaces
	if err := p.sendNamespaces(); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "sending namespaces to init process")
	}

	// exec process会在启动后阻塞，直至收到config
	if err = p.sendConfig(); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "sending init config to init process")
	}
	// parent 写完就关
	if err = p.parentConfigPipe.Close(); err != nil {
		logrus.Errorf("closing parent pipe failed: %s", err.Error())
	}
	return p.wait()
}

func (p *ParentExecProcess) detach() bool {
	return p.process.Detach
}

func (p *ParentExecProcess) pid() int {
	return p.execProcessCmd.Process.Pid
}

func (p *ParentExecProcess) startTime() (uint64, error) {
	stat, err := proc.GetProcessStat(p.pid())
	if err != nil {
		return 0, err
	}
	return stat.StartTime, err
}

func (p *ParentExecProcess) terminate() error {
	if p.execProcessCmd.Process == nil {
		logrus.Warnf("exec process is nil, cant be terminated")
		return nil
	}
	err := p.execProcessCmd.Process.Kill()
	if err := p.wait(); err == nil {
		return err
	}
	return err
}

func (p *ParentExecProcess) wait() error {
	logrus.Infof("starting to wait exec process exit")
	err := p.execProcessCmd.Wait()
	if err != nil {
		return err
	}
	logrus.Infof("wait exec process exit complete")
	return nil
}

func (p *ParentExecProcess) signal(os.Signal) error {
	panic("implement me")
}

func (p *ParentExecProcess) sendNamespaces() error {
	initProcessPid, err := p.container.loadContainerInitProcessPid()
	if err != nil {
		return err
	}
	var namespacePaths []string
	for _, ns := range p.container.config.Namespaces {
		namespacePaths = append(namespacePaths, ns.GetPath(initProcessPid))
	}
	logrus.Infof("sending config: %#v", namespacePaths)
	data := []byte(strings.Join(namespacePaths, ","))
	lenInBytes, err := util.Int32ToBytes(int32(len(data)))
	if err != nil {
		return err
	}
	if _, err := p.parentConfigPipe.Write(lenInBytes); err != nil {
		return err
	}
	if _, err := p.parentConfigPipe.Write(data); err != nil {
		return err
	}
	return nil
}

func (p *ParentExecProcess) sendConfig() error {
	return sendConfig(p.container.config, *p.process, p.container.id, p.parentConfigPipe)
}
