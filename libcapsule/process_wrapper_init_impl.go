package libcapsule

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/cgroups"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"github.com/songxinjianqwe/rune/libcapsule/util/system"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"syscall"
)

func NewInitProcessWrapper(process *Process, cmd *exec.Cmd, parentPipe *os.File, childPipe *os.File, c *LinuxContainerImpl) ProcessWrapper {
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", EnvInitializerType, string(StandardInitializer)))
	logrus.Infof("new init process wrapper...")
	return &InitProcessWrapperImpl{
		initProcessCmd:    cmd,
		parentPipe:        parentPipe,
		childPipe:         childPipe,
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
	parentPipe        *os.File
	childPipe         *os.File
	container         *LinuxContainerImpl
	process           *Process
	cgroupManger      *cgroups.CgroupManager
	sharePidNamespace bool
}

type InitConfig struct {
	ContainerConfig configc.Config
	ProcessConfig   Process
}

func (p *InitProcessWrapperImpl) start() error {
	logrus.Infof("InitProcessWrapperImpl starting...")
	err := p.initProcessCmd.Start()
	p.childPipe.Close()
	if err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "starting init process command")
	}
	if err := p.container.cgroupManager.Apply(p.pid()); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "applying cgroup configuration for process")
	}
	defer func() {
		if err != nil {
			p.container.cgroupManager.Destroy()
		}
	}()
	if err = p.createNetworkInterfaces(); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "creating network interfaces")
	}
	// init process会在启动后阻塞，直至收到config
	if err = p.sendConfig(); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "sending config to init process")
	}
	if err = p.parentPipe.Close(); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "closing parent pipe")
	}
	state, err := p.wait()
	if err != nil {
		logrus.Errorf("waiting init process cmd error:%v, %s", state, err.Error())
	}
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
	logrus.Infof("wait init process exit complete")
	if err != nil {
		return p.initProcessCmd.ProcessState, err
	}
	// we should kill all processes in cgroup when init is died if we use host PID namespace
	if p.sharePidNamespace {
		system.SignalAllProcesses(*p.cgroupManger, unix.SIGKILL)
	}
	return p.initProcessCmd.ProcessState, nil
}

func (p *InitProcessWrapperImpl) startTime() (uint64, error) {
	stat, err := system.GetProcessStat(p.pid())
	if err != nil {
		return -1, err
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
	logrus.Infof("sending config:%#v", initConfig)
	bytes, err := json.Marshal(initConfig)
	if err != nil {
		return err
	}
	_, err = p.parentPipe.WriteString(string(bytes))
	return err
}
