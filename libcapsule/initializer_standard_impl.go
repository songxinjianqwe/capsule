package libcapsule

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

type InitializerStandardImpl struct {
	config     *InitConfig
	configPipe *os.File
	parentPid  int
}

// **************************************************************************************************
// public
// **************************************************************************************************

/**
容器初始化
*/
func (initializer *InitializerStandardImpl) Init() error {
	logrus.WithField("init", true).Infof("InitializerStandardImpl create to Init()")
	// 初始化网络
	if err := initializer.setUpNetwork(); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "init config/set up network")
	}

	// 初始化路由
	if err := initializer.setUpRoute(); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "init config/set up route")
	}

	// 初始化rootfs
	if err := initializer.prepareRootfs(); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "init config/prepare rootfs")
	}

	// 初始化hostname
	if hostname := initializer.config.ContainerConfig.Hostname; hostname != "" {
		logrus.WithField("init", true).Infof("setting hostname: %s", hostname)
		if err := unix.Sethostname([]byte(hostname)); err != nil {
			return util.NewGenericErrorWithContext(err, util.SystemError, "init config/set hostname")
		}
	}

	// 初始化环境变量
	for key, value := range initializer.config.ContainerConfig.Sysctl {
		if err := writeSystemProperty(key, value); err != nil {
			return util.NewGenericErrorWithContext(err, util.SystemError, fmt.Sprintf("write sysctl key %s", key))
		}
	}

	// 初始化namespace
	if err := initializer.finalizeNamespace(); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "init config/finalize namespace")
	}

	// look path 可以在系统的PATH里面寻找命令的绝对路径
	name, err := exec.LookPath(initializer.config.ProcessConfig.Args[0])
	if err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "init config/look path cmd")
	}
	logrus.WithField("init", true).Infof("look path: %s", name)

	logrus.WithField("init", true).Infof("sync parent ready...")

	// 告诉parent，init process已经初始化完毕，马上要执行命令了
	if err := unix.Kill(initializer.parentPid, syscall.SIGUSR1); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "init config/sync parent ready")
	}

	// 等待parent给一个继续执行命令，即exec的信号
	logrus.WithField("init", true).Info("start waiting parent continue(SIGUSR2) signal...")
	receivedChan := make(chan os.Signal, 1)
	signal.Notify(receivedChan, syscall.SIGUSR2)
	<-receivedChan
	logrus.WithField("init", true).Info("received parent continue(SIGUSR2) signal")

	logrus.WithField("init", true).Info("execute real command and cover rune init config")
	// syscall.Exec与cmd.Start不同，后者是启动一个新的进程来执行命令
	// 而前者会在覆盖当前进程的镜像、数据、堆栈等信息，包括PID。
	logrus.WithField("init", true).Infof("syscall.Exec(name: %s, args: %v, env: %v)...", name, initializer.config.ProcessConfig.Args, os.Environ())
	// 在执行这条命令后，当前进程的命令会变化，但pid不变，同时parent进程死掉，当前进程的父进程变为pid=1的进程
	// 问题是在输入任何指令后，当前进程会立即结束，并且ssh结束/当前登录用户的会话结束
	if err := syscall.Exec(name, initializer.config.ProcessConfig.Args, os.Environ()); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "start user config")
	}
	return nil
}

// **************************************************************************************************
// private
// **************************************************************************************************

func (initializer *InitializerStandardImpl) setUpNetwork() error {
	logrus.WithField("init", true).Info("setting up network...")
	return nil
}

func (initializer *InitializerStandardImpl) setUpRoute() error {
	logrus.WithField("init", true).Info("setting up route...")
	return nil
}

func (initializer *InitializerStandardImpl) prepareRootfs() error {
	logrus.WithField("init", true).Info("preparing rootfs...")
	// 挂载
	for _, m := range initializer.config.ContainerConfig.Mounts {
		if err := mountToRootfs(m, initializer.config.ContainerConfig.Rootfs); err != nil {
			return util.NewGenericErrorWithContext(err, util.SystemError, fmt.Sprintf("mounting %q to rootfs %q at %q", m.Source, initializer.config.ContainerConfig.Rootfs, m.Destination))
		}
	}
	return nil
}

func (initializer *InitializerStandardImpl) finalizeNamespace() error {
	logrus.WithField("init", true).Info("finalizing namespace...")
	cwd := initializer.config.ProcessConfig.Cwd
	if cwd != "" {
		logrus.WithField("init", true).Info("changing dir to cwd: %s", cwd)
		if err := os.Chdir(cwd); err != nil {
			return fmt.Errorf("chdir to cwd (%q) set in config.json failed: %v", cwd, err)
		}
	}
	return nil
}

// **************************************************************************************************
// util
// **************************************************************************************************

func writeSystemProperty(key string, value string) error {
	logrus.WithField("init", true).Infof("write system property:key:%s, value:%s", key, value)
	return nil
}

func maskPath(path string, labels []string) error {
	logrus.WithField("init", true).Infof("mask path:path:%s, labels:%v", path, labels)
	return nil
}

func readonlyPath(path string) error {
	logrus.WithField("init", true).Infof("make path read only:path:%s", path)
	return nil
}
