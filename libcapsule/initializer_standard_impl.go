package libcapsule

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"syscall"
)

type InitializerStandardImpl struct {
	config     *InitConfig
	childPipe  *os.File
	execFifoFd int
}

// **************************************************************************************************
// public
// **************************************************************************************************

/**
容器初始化
*/
func (initializer *InitializerStandardImpl) Init() error {
	logrus.Infof("InitializerStandardImpl start to Init()")
	if err := initializer.setUpNetwork(); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "init process/set up network")
	}
	if err := initializer.setUpRoute(); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "init process/set up route")
	}
	if err := initializer.prepareRootfs(); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "init process/prepare rootfs")
	}
	if hostname := initializer.config.ContainerConfig.Hostname; hostname != "" {
		logrus.Info("set hostname")
		if err := unix.Sethostname([]byte(hostname)); err != nil {
			return util.NewGenericErrorWithInfo(err, util.SystemError, "init process/set hostname")
		}
	}
	for key, value := range initializer.config.ContainerConfig.Sysctl {
		if err := writeSystemProperty(key, value); err != nil {
			return util.NewGenericErrorWithInfo(err, util.SystemError, fmt.Sprintf("write sysctl key %s", key))
		}
	}
	for _, path := range initializer.config.ContainerConfig.ReadonlyPaths {
		if err := readonlyPath(path); err != nil {
			return util.NewGenericErrorWithInfo(err, util.SystemError, "init process/set path read only")
		}
	}
	for _, path := range initializer.config.ContainerConfig.MaskPaths {
		if err := maskPath(path, initializer.config.ContainerConfig.Labels); err != nil {
			return util.NewGenericErrorWithInfo(err, util.SystemError, "init process/set path mask")
		}
	}
	if err := initializer.finalizeNamespace(); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "init process/finalize namespace")
	}
	// look path 可以在系统的PATH里面寻找命令的绝对路径
	name, err := exec.LookPath(initializer.config.ProcessConfig.Args[0])
	if err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "init process/look path cmd")
	}
	logrus.Infof("look path: %s", name)
	if err := initializer.childPipe.Close(); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "init process/close child pipe")
	}
	logrus.Info("open exec fifo")
	fifo, err := os.OpenFile(fmt.Sprintf("/prod/self/fd/%d", initializer.execFifoFd), os.O_WRONLY, 0)
	if err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "open exec fifo")
	}
	logrus.Info("write 0 to exec fifo and block here")
	// block here
	if _, err := fifo.Write([]byte{0}); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "write 0 to exec fifo")
	}
	logrus.Info("close exec fifo")
	if err := fifo.Close(); err != nil {
		fmt.Printf("close fifo error: %s", err.Error())
	}
	logrus.Info("execute real command and cover rune init process")
	// syscall.Exec与cmd.Start不同，后者是启动一个新的进程来执行命令
	// 而前者会在覆盖当前进程的镜像、数据、堆栈等信息，包括PID。
	if err := syscall.Exec(name, initializer.config.ProcessConfig.Args[0:], os.Environ()); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "exec user process")
	}
	logrus.Info("execute real command complete")
	return nil
}

// **************************************************************************************************
// private
// **************************************************************************************************

func (initializer *InitializerStandardImpl) setUpNetwork() error {
	logrus.Info("setup network")
	return nil
}

func (initializer *InitializerStandardImpl) setUpRoute() error {
	logrus.Info("setup route")
	return nil
}

func (initializer *InitializerStandardImpl) prepareRootfs() error {
	logrus.Info("prepare rootfs")
	return nil
}

func (initializer *InitializerStandardImpl) finalizeNamespace() error {
	logrus.Info("finalize namespace")
	return nil
}

// **************************************************************************************************
// util
// **************************************************************************************************

func writeSystemProperty(key string, value string) error {
	logrus.Infof("write system property:key:%s, value:%s", key, value)
	return nil
}

func maskPath(path string, labels []string) error {
	logrus.Infof("mask path:path:%s, labels:%v", path, labels)
	return nil
}

func readonlyPath(path string) error {
	logrus.Infof("make path read only:path:%s", path)
	return nil
}
