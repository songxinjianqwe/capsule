package libcapsule

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"github.com/songxinjianqwe/rune/libcapsule/util/system"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func NewParentInitProcess(process *Process, cmd *exec.Cmd, parentConfigPipe *os.File, c *LinuxContainer) ParentProcess {
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", EnvInitializerType, string(StandardInitializer)))
	logrus.Infof("new init process wrapper...")
	return &ParentInitProcess{
		initProcessCmd:   cmd,
		parentConfigPipe: parentConfigPipe,
		container:        c,
		process:          process,
	}
}

/**
ParentProcess接口的实现类，包裹了InitProcess，它返回的进程信息均为容器Init进程的信息
*/
type ParentInitProcess struct {
	initProcessCmd   *exec.Cmd
	parentConfigPipe *os.File
	parentExecPipe   *os.File
	container        *LinuxContainer
	process          *Process
}

type InitConfig struct {
	ContainerConfig configc.Config `json:"container_config"`
	ProcessConfig   Process        `json:"process_config"`
}

func (p *ParentInitProcess) detach() bool {
	return p.process.Detach
}

func (p *ParentInitProcess) start() error {
	logrus.Infof("ParentInitProcess starting...")
	err := p.initProcessCmd.Start()
	if err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "starting init process command")
	}
	if err := p.container.cgroupManager.Apply(p.pid()); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "applying cgroup configuration for process")
	}
	defer func() {
		if err != nil {
			p.container.cgroupManager.Destroy()
		}
	}()

	if err = p.createNetworkInterfaces(); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "creating network interfaces")
	}

	// init process会在启动后阻塞，直至收到config
	if err = p.sendConfig(); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "sending config to init config")
	}
	// parent 写完就关
	if err = p.parentConfigPipe.Close(); err != nil {
		logrus.Errorf("closing parent pipe failed: %s", err.Error())
	}

	if err := p.setupResourceLimits(); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "setting rlimits for ready config")
	}

	// 等待init process到达在初始化之后，执行命令之前的状态
	// 使用SIGUSR1信号
	logrus.Info("start waiting init process ready(SIGUSR1) or fail(SIGCHLD) signal...")
	receivedChan := make(chan os.Signal, 1)
	signal.Notify(receivedChan, syscall.SIGUSR1, syscall.SIGCHLD)
	sig := <-receivedChan
	if sig == syscall.SIGUSR1 {
		logrus.Info("received SIGUSR1 signal")
	} else if sig == syscall.SIGCHLD {
		logrus.Errorf("received SIGCHLD signal")
		return fmt.Errorf("init process init failed")
	}
	return nil
}

func (p *ParentInitProcess) pid() int {
	return p.initProcessCmd.Process.Pid
}

func (p *ParentInitProcess) terminate() error {
	if p.initProcessCmd.Process == nil {
		return nil
	}
	err := p.initProcessCmd.Process.Kill()
	if err := p.wait(); err == nil {
		return err
	}
	return err
}

func (p *ParentInitProcess) wait() error {
	logrus.Infof("starting to wait init config exit")
	err := p.initProcessCmd.Wait()
	if err != nil {
		return err
	}
	logrus.Infof("wait init config exit complete")
	return nil
}

func (p *ParentInitProcess) startTime() (uint64, error) {
	stat, err := system.GetProcessStat(p.pid())
	if err != nil {
		return 0, err
	}
	return stat.StartTime, err
}

func (p *ParentInitProcess) signal(sig os.Signal) error {
	s, ok := sig.(syscall.Signal)
	if !ok {
		return util.NewGenericError(fmt.Errorf("os: unsupported signal type:%v", sig), util.SystemError)
	}
	return unix.Kill(p.pid(), s)
}

// ******************************************************************************************************
// biz methods
// ******************************************************************************************************

func (p *ParentInitProcess) createNetworkInterfaces() error {
	logrus.Infof("creating network interfaces")
	return nil
}

func (p *ParentInitProcess) sendConfig() error {
	initConfig := &InitConfig{
		ContainerConfig: p.container.config,
		ProcessConfig:   *p.process,
	}
	// remove unnecessary fields, or init config will unmarshal it failed
	initConfig.ProcessConfig.Stdin = nil
	initConfig.ProcessConfig.Stdout = nil
	initConfig.ProcessConfig.Stderr = nil
	initConfig.ProcessConfig.ExtraFiles = nil

	logrus.Infof("sending config: %#v", initConfig)
	bytes, err := json.Marshal(initConfig)
	if err != nil {
		return err
	}
	_, err = p.parentConfigPipe.WriteString(string(bytes))
	return err
}

func (p *ParentInitProcess) setupResourceLimits() error {
	logrus.Infof("setting up resource limits")
	return nil
}
