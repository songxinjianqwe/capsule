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
	logrus.Infof("new parent process...")
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
	ID              string         `json:"id"`
}

func (p *ParentInitProcess) detach() bool {
	return p.process.Detach
}

func (p *ParentInitProcess) start() (err error) {
	logrus.Infof("ParentInitProcess starting...")
	defer func() {
		// 如果start方法出现任何异常，则必须销毁cgroup manager
		if err != nil {
			logrus.Warnf("parent process init failed, destroying cgroup manager...")
			destroyErr := p.container.cgroupManager.Destroy()
			if destroyErr != nil {
				logrus.Warnf("destroy failed, cause: %s", destroyErr.Error())
			}
		}
	}()

	err = p.initProcessCmd.Start()
	if err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "starting init process command")
	}
	logrus.Infof("init process started, INIT_PROCESS_PID: [%d]", p.pid())

	util.WaitUserEnterGo()

	// 设置cgroup config
	if err = p.container.cgroupManager.SetConfig(p.container.config.CgroupConfig); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "setting cgroup config for procHooks process")
	}
	// 将pid加入到cgroup set中
	if err = p.container.cgroupManager.JoinCgroupSet(p.pid()); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "applying cgroup configuration for process")
	}
	util.PrintSubsystemPids("memory", p.container.id, "after cgroup manager init", false)

	// 创建网络接口，比如bridge
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

	// 等待init process到达在初始化之后，执行命令之前的状态
	// 使用SIGUSR1信号
	logrus.Info("start waiting init process ready(SIGUSR1) or fail(SIGCHLD) signal...")
	receivedChan := make(chan os.Signal, 1)
	signal.Notify(receivedChan, syscall.SIGUSR1, syscall.SIGCHLD)
	sig := <-receivedChan
	if sig == syscall.SIGUSR1 {
		logrus.Info("received SIGUSR1 signal")
		util.PrintSubsystemPids("memory", p.container.id, "after init process ready", false)
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
	logrus.Infof("starting to wait init process exit")
	err := p.initProcessCmd.Wait()
	if err != nil {
		return err
	}
	logrus.Infof("wait init process exit complete")
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
		ID:              p.container.id,
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
