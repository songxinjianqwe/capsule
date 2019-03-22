package libcapsule

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
	"github.com/songxinjianqwe/capsule/libcapsule/util"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
	"github.com/songxinjianqwe/capsule/libcapsule/util/proc"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type ParentProcessStartHook func(*ParentAbstractProcess) error

type ParentAbstractProcess struct {
	// init or exec process cmd
	processCmd       *exec.Cmd
	parentConfigPipe *os.File
	container        *LinuxContainer
	process          *Process
	cloneFlags       uintptr
	namespacePathMap map[configs.NamespaceType]string
	// 模板方法模式
	startHook ParentProcessStartHook
}

func (p *ParentAbstractProcess) pid() int {
	return p.processCmd.Process.Pid
}

func (p *ParentAbstractProcess) terminate() error {
	if p.processCmd.Process == nil {
		return nil
	}
	err := p.processCmd.Process.Kill()
	if err := p.wait(); err == nil {
		return err
	}
	return err
}

func (p *ParentAbstractProcess) wait() error {
	logrus.Infof("starting to wait init process exit")
	err := p.processCmd.Wait()
	if err != nil {
		return err
	}
	logrus.Infof("wait init process exit complete")
	return nil
}

func (p *ParentAbstractProcess) startTime() (uint64, error) {
	stat, err := proc.GetProcessStat(p.pid())
	if err != nil {
		return 0, err
	}
	return stat.StartTime, err
}

func (p *ParentAbstractProcess) signal(sig os.Signal) error {
	s, ok := sig.(syscall.Signal)
	if !ok {
		return exception.NewGenericError(fmt.Errorf("os: unsupported signal type:%v", sig), exception.SignalError)
	}
	return unix.Kill(p.pid(), s)
}

func (p *ParentAbstractProcess) detach() bool {
	return p.process.Detach
}

/*
模板方法模式
*/
func (p *ParentAbstractProcess) start() error {
	logrus.Infof("ParentAbstractProcess starting...")
	if err := p.processCmd.Start(); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.CmdStartError, "starting init/exec process command")
	}
	logrus.Infof("INIT/EXEC PROCESS STARTED, PID: %d", p.pid())
	if err := p.sendNamespaces(); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.PipeError, "sending namespacePathMap to init/exec process")
	}
	if err := p.sendCloneFlags(); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.PipeError, "sending clone flags to init/exec process")
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
	p.processCmd.Process = process
	return p.startHook(p)
}

func (p *ParentAbstractProcess) sendCloneFlags() error {
	// 这里只传没有Path(即需要新建)的NS的clone flags
	logrus.Infof("sending clone flags: %d", p.cloneFlags)
	bytes, err := util.Int32ToBytes(int32(p.cloneFlags))
	if err != nil {
		return err
	}
	if _, err := p.parentConfigPipe.Write(bytes); err != nil {
		return err
	}
	return nil
}

func (p *ParentAbstractProcess) sendNamespaces() error {
	var namespacePaths []string
	// order
	// mnt必须在最后
	for _, ns := range configs.AllNamespaceTypes() {
		if path, exist := p.namespacePathMap[ns]; exist {
			namespacePaths = append(namespacePaths, path)
		}
	}
	logrus.Infof("sending namespacePathMap: %#v", namespacePaths)
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

func (p *ParentAbstractProcess) sendConfigAndClosePipe() error {
	initConfig := &InitExecConfig{
		ContainerConfig: p.container.config,
		ProcessConfig:   *p.process,
		ID:              p.container.id,
	}
	logrus.Infof("sending config: %#v", initConfig)
	bytes, err := json.Marshal(initConfig)
	if err != nil {
		return err
	}
	_, err = p.parentConfigPipe.WriteString(string(bytes))
	if err != nil {
		return err
	}
	// parent 写完就关
	if err := p.parentConfigPipe.Close(); err != nil {
		logrus.Errorf("closing parent pipe failed: %s", err.Error())
	}
	return nil
}
