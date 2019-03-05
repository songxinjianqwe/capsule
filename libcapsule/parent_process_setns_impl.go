package libcapsule

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/util"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
	"os"
	"os/exec"
	"strings"
)

/**
ParentProcess接口的实现类，包裹了SetnsProcess
*/
type ParentSetnsProcess struct {
	execProcessCmd   *exec.Cmd
	parentConfigPipe *os.File
	parentExecPipe   *os.File
	container        *LinuxContainer
	process          *Process
}

func (p *ParentSetnsProcess) start() error {
	logrus.Infof("ParentSetnsProcess starting...")
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

	return nil
}

func (p *ParentSetnsProcess) detach() bool {
	return p.process.Detach
}

func (p *ParentSetnsProcess) pid() int {
	return p.execProcessCmd.Process.Pid
}

func (p *ParentSetnsProcess) terminate() error {
	if p.execProcessCmd.Process == nil {
		return nil
	}
	err := p.execProcessCmd.Process.Kill()
	if err := p.wait(); err == nil {
		return err
	}
	return err
}

func (p *ParentSetnsProcess) wait() error {
	panic("implement me")
}

func (p *ParentSetnsProcess) startTime() (uint64, error) {
	panic("implement me")
}

func (p *ParentSetnsProcess) signal(os.Signal) error {
	panic("implement me")
}

func (p *ParentSetnsProcess) sendConfig() error {
	initConfig := &InitConfig{
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
	return err
}

func (p *ParentSetnsProcess) sendNamespaces() error {
	var namespacePaths []string
	for _, ns := range p.container.config.Namespaces {
		namespacePaths = append(namespacePaths, ns.GetPath(p.pid()))
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
