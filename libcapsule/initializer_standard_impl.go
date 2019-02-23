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
	logrus.WithField("init", true).Infof("InitializerStandardImpl start to Init()")
	if err := initializer.setUpNetwork(); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "init process/set up network")
	}
	if err := initializer.setUpRoute(); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "init process/set up route")
	}

	if err := initializer.prepareRootfs(); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "init process/prepare rootfs")
	}

	if hostname := initializer.config.ContainerConfig.Hostname; hostname != "" {
		logrus.WithField("init", true).Infof("setting hostname: %s", hostname)
		if err := unix.Sethostname([]byte(hostname)); err != nil {
			return util.NewGenericErrorWithContext(err, util.SystemError, "init process/set hostname")
		}
	}

	for key, value := range initializer.config.ContainerConfig.Sysctl {
		if err := writeSystemProperty(key, value); err != nil {
			return util.NewGenericErrorWithContext(err, util.SystemError, fmt.Sprintf("write sysctl key %s", key))
		}
	}

	for _, path := range initializer.config.ContainerConfig.ReadonlyPaths {
		if err := readonlyPath(path); err != nil {
			return util.NewGenericErrorWithContext(err, util.SystemError, "init process/set path read only")
		}
	}

	for _, path := range initializer.config.ContainerConfig.MaskPaths {
		if err := maskPath(path, initializer.config.ContainerConfig.Labels); err != nil {
			return util.NewGenericErrorWithContext(err, util.SystemError, "init process/set path mask")
		}
	}

	if err := initializer.finalizeNamespace(); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "init process/finalize namespace")
	}

	// look path 可以在系统的PATH里面寻找命令的绝对路径
	name, err := exec.LookPath(initializer.config.ProcessConfig.Args[0])
	if err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "init process/look path cmd")
	}
	logrus.WithField("init", true).Infof("look path: %s", name)

	logrus.WithField("init", true).Infof("sync parent ready...")
	// 告诉parent，init process已经初始化完毕，马上要执行命令了
	if err := unix.Kill(initializer.parentPid, syscall.SIGUSR1); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "init process/sync parent ready")
	}

	// 等待parent给一个继续执行命令，即exec的信号
	logrus.WithField("init", true).Info("start to wait parent continue(SIGUSR2) signal...")
	receivedChan := make(chan os.Signal, 1)
	signal.Notify(receivedChan, syscall.SIGUSR2)
	<-receivedChan
	logrus.WithField("init", true).Info("received SIGUSR2 signal")

	logrus.WithField("init", true).Info("execute real command and cover rune init process")
	// syscall.Exec与cmd.Start不同，后者是启动一个新的进程来执行命令
	// 而前者会在覆盖当前进程的镜像、数据、堆栈等信息，包括PID。
	if err := syscall.Exec(name, initializer.config.ProcessConfig.Args, os.Environ()); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "exec user process")
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
	return nil
}

func (initializer *InitializerStandardImpl) finalizeNamespace() error {
	logrus.WithField("init", true).Info("finalizing namespace...")
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
