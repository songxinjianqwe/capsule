package libcapsule

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
	"github.com/songxinjianqwe/capsule/libcapsule/util"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
	"github.com/songxinjianqwe/capsule/libcapsule/util/proc"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"syscall"
)

type InitConfig struct {
	ContainerConfig configs.ContainerConfig `json:"container_config"`
	ProcessConfig   Process                 `json:"process_config"`
	ID              string                  `json:"id"`
}

/**
ParentProcess接口的实现类，包裹了InitProcess，它返回的进程信息均为容器Init进程的信息
*/
type ParentInitProcess struct {
	initProcessCmd   *exec.Cmd
	parentConfigPipe *os.File
	container        *LinuxContainer
	process          *Process
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
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "starting init process command")
	}
	logrus.Infof("init process started, INIT_PROCESS_PID: [%d]", p.pid())

	// 将pid加入到cgroup set中
	if err = p.container.cgroupManager.JoinCgroupSet(p.pid()); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "applying cgroup configuration for process")
	}
	util.PrintSubsystemPids("memory", p.container.id, "after cgroup manager init", false)

	// 设置cgroup config
	if err = p.container.cgroupManager.SetConfig(p.container.config.Cgroup); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "setting cgroup config for procHooks process")
	}

	// 创建网络接口，比如bridge
	if err = p.createNetworkInterfaces(); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "creating network interfaces")
	}

	// init process会在启动后阻塞，直至收到config
	if err = p.sendConfig(); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "sending config to init process")
	}

	// parent 写完就关
	if err = p.parentConfigPipe.Close(); err != nil {
		logrus.Errorf("closing parent pipe failed: %s", err.Error())
	}

	// 等待init process到达在初始化之后，执行命令之前的状态
	// 使用SIGUSR1信号
	logrus.Info("start waiting init process ready(SIGUSR1) or fail(SIGCHLD) signal...")
	sig := util.WaitSignal(syscall.SIGUSR1, syscall.SIGCHLD)
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
	logrus.Infof("starting to kill init process")
	if err := p.initProcessCmd.Process.Kill(); err != nil {
		return err
	}
	if err := p.wait(); err != nil {
		return err
	}
	return nil
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
	stat, err := proc.GetProcessStat(p.pid())
	if err != nil {
		return 0, err
	}
	return stat.StartTime, err
}

func (p *ParentInitProcess) signal(sig os.Signal) error {
	s, ok := sig.(syscall.Signal)
	if !ok {
		return exception.NewGenericError(fmt.Errorf("os: unsupported signal type:%v", sig), exception.SystemError)
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
	return sendConfig(p.container.config, *p.process, p.container.id, p.parentConfigPipe)
}
