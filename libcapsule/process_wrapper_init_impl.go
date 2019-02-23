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

func NewInitProcessWrapper(process *Process, cmd *exec.Cmd, parentConfigPipe *os.File, c *LinuxContainerImpl) ProcessWrapper {
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", EnvInitializerType, string(StandardInitializer)))
	logrus.Infof("new init process wrapper...")
	return &InitProcessWrapperImpl{
		initProcessCmd:    cmd,
		parentConfigPipe:  parentConfigPipe,
		container:         c,
		process:           process,
		sharePidNamespace: c.config.Namespaces.Contains(configc.NEWPID),
	}
}

/**
ProcessWrapper接口的实现类，包裹了InitProcess，它返回的进程信息均为容器Init进程的信息
*/
type InitProcessWrapperImpl struct {
	initProcessCmd    *exec.Cmd
	parentConfigPipe  *os.File
	parentExecPipe    *os.File
	container         *LinuxContainerImpl
	process           *Process
	sharePidNamespace bool
}

type InitConfig struct {
	ContainerConfig configc.Config
	ProcessConfig   Process
}

func (p *InitProcessWrapperImpl) start() error {
	logrus.Infof("InitProcessWrapperImpl starting...")
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
		return util.NewGenericErrorWithContext(err, util.SystemError, "sending config to init process")
	}
	// parent 写完就关
	if err = p.parentConfigPipe.Close(); err != nil {
		logrus.Errorf("closing parent pipe failed: %s", err.Error())
	}
	// set rlimits, this has to be done here because we lose permissions
	// to raise the limits once we enter a user-namespace
	if err := p.setupResourceLimits(); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "setting rlimits for ready process")
	}
	// 等待init process到达在初始化之后，执行命令之前的状态
	// 使用SIGUSR1信号
	logrus.WithField("init", true).Info("start to wait init process ready(SIGUSR1) signal...")
	receivedChan := make(chan os.Signal, 1)
	signal.Notify(receivedChan, syscall.SIGUSR1)
	<-receivedChan
	logrus.WithField("init", true).Info("received SIGUSR1 signal")
	return nil
}

func (p *InitProcessWrapperImpl) pid() int {
	return p.initProcessCmd.Process.Pid
}

func (p *InitProcessWrapperImpl) terminate() error {
	panic("implement me")
}

func (p *InitProcessWrapperImpl) wait() (*os.ProcessState, error) {
	logrus.Infof("starting to wait init process exit")
	err := p.initProcessCmd.Wait()
	if err != nil {
		return p.initProcessCmd.ProcessState, err
	}
	logrus.Infof("wait init process exit complete")
	// we should kill all processes in cgroup when init is died if we use host PID namespace
	if p.sharePidNamespace {
		system.SignalAllProcesses(p.container.cgroupManager, unix.SIGKILL)
	}
	return p.initProcessCmd.ProcessState, nil
}

func (p *InitProcessWrapperImpl) startTime() (uint64, error) {
	stat, err := system.GetProcessStat(p.pid())
	if err != nil {
		return 0, err
	}
	return stat.StartTime, err
}

func (p *InitProcessWrapperImpl) signal(sig os.Signal) error {
	s, ok := sig.(syscall.Signal)
	if !ok {
		return util.NewGenericError(fmt.Errorf("os: unsupported signal type:%v", sig), util.SystemError)
	}
	return unix.Kill(p.pid(), s)
}

// ******************************************************************************************************
// biz methods
// ******************************************************************************************************

func (p *InitProcessWrapperImpl) createNetworkInterfaces() error {
	logrus.Infof("creating network interfaces")
	return nil
}

func (p *InitProcessWrapperImpl) sendConfig() error {
	initConfig := &InitConfig{
		ContainerConfig: p.container.config,
		ProcessConfig:   *p.process,
	}
	// remove unnecessary fields, or init process will unmarshal it failed
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

func (p *InitProcessWrapperImpl) setupResourceLimits() error {
	logrus.Infof("setting up resource limits")
	return nil
}
