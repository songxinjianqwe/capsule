package libcapsule

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
	"github.com/songxinjianqwe/capsule/libcapsule/util"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
	"github.com/songxinjianqwe/capsule/libcapsule/util/proc"
	"os"
	"os/exec"
	"strings"
)

/*
ParentProcess接口的实现类，包裹了ExecProcess
*/
type ParentExecProcess struct {
	execProcessCmd   *exec.Cmd
	parentConfigPipe *os.File
	container        *LinuxContainer
	process          *Process
}

/*
对于Exec来说，start返回后，非daemon的进程已经结束了。
*/
func (p *ParentExecProcess) start() error {
	logrus.Infof("ParentExecProcess starting...")
	err := p.execProcessCmd.Start()
	if err != nil {
		return exception.NewGenericErrorWithContext(err, exception.CmdStartError, "starting init process command")
	}
	logrus.Infof("exec process started, EXEC_PROCESS_PID: [%d]", p.pid())

	// exec process会在启动后阻塞，直至收到namespaces
	if err := p.sendNamespaces(); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.PipeError, "sending namespaces to init process")
	}

	childPid, err := util.ReadIntFromFile(p.parentConfigPipe)
	logrus.Infof("read child pid from parent pipe: %d", childPid)
	if err != nil {
		return exception.NewGenericErrorWithContext(err, exception.PipeError, "reading child pid")
	}
	process, err := os.FindProcess(childPid)
	if err != nil {
		return err
	}
	logrus.Infof("find new child process: %#v", process)
	p.execProcessCmd.Process = process

	// 如果启用了cgroups，那么将exec进程的pid也加进去
	if len(p.container.cgroupManager.GetPaths()) > 0 {
		if err := p.container.cgroupManager.JoinCgroupSet(p.pid()); err != nil {
			return exception.NewGenericErrorWithContext(err, exception.CgroupsError, fmt.Sprintf("adding pid %d to cgroups", p.pid()))
		}
	}

	// exec process会在启动后阻塞，直至收到config
	if err = p.sendConfig(); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.PipeError, "sending config to init process")
	}

	// parent 写完就关
	if err = p.parentConfigPipe.Close(); err != nil {
		logrus.Errorf("closing parent pipe failed: %s", err.Error())
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
	var namespacePaths []string
	state, err := p.container.currentState()
	if err != nil {
		return err
	}
	// order mnt必须在最后
	for _, ns := range configs.AllNamespaceTypes() {
		if path, exist := state.NamespacePaths[ns]; exist {
			namespacePaths = append(namespacePaths, path)
		}
	}
	logrus.Infof("sending namespaces: %#v", namespacePaths)
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

func (p *ParentExecProcess) sendCommand() error {
	argsString := strings.Join(p.process.Args, " ")
	lenInBytes, err := util.Int32ToBytes(int32(len(argsString)))
	if err != nil {
		return err
	}
	logrus.Infof("write length of commands: %v", lenInBytes)
	if _, err := p.parentConfigPipe.Write(lenInBytes); err != nil {
		return err
	}
	if _, err := p.parentConfigPipe.WriteString(argsString); err != nil {
		return err
	}
	return nil
}

func (p *ParentExecProcess) sendConfig() error {
	return sendConfig(p.container.config, *p.process, p.container.id, p.parentConfigPipe)
}
