package libcapsule

import (
	"fmt"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"syscall"
)

func NewInitializer(config *InitConfig, childPipe *os.File, execFifoFd int) Initializer {
	return &InitializerImpl{
		config:     config,
		childPipe:  childPipe,
		execFifoFd: execFifoFd,
	}
}

type InitializerImpl struct {
	config     *InitConfig
	childPipe  *os.File
	execFifoFd int
}

// **************************************************************************************************
// public
// **************************************************************************************************

/**
容器init进程初始化，即容器初始化
*/
func (initializer *InitializerImpl) Init() error {
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
	if err := initializer.childPipe.Close(); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "init process/close child pipe")
	}
	fifo, err := os.OpenFile(fmt.Sprintf("/prod/self/fd/%d", initializer.execFifoFd), os.O_WRONLY, 0)
	if err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "open exec fifo")
	}
	// hang
	if _, err := fifo.Write([]byte{0}); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "write 0 to exec fifo")
	}
	fifo.Close()
	// syscall.Exec与cmd.Start不同，后者是启动一个新的进程来执行命令
	// 而前者会在覆盖当前进程的镜像、数据、堆栈等信息，包括PID。
	if err := syscall.Exec(name, initializer.config.ProcessConfig.Args[0:], os.Environ()); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "exec user process")
	}
	return nil
}

// **************************************************************************************************
// private
// **************************************************************************************************

func (initializer *InitializerImpl) setUpNetwork() error {
	return nil
}

func (initializer *InitializerImpl) setUpRoute() error {
	return nil
}

func (initializer *InitializerImpl) prepareRootfs() error {
	return nil
}

func (initializer *InitializerImpl) finalizeNamespace() error {
	return nil
}

// **************************************************************************************************
// util
// **************************************************************************************************

func writeSystemProperty(key string, value string) error {
	return nil
}

func maskPath(path string, labels []string) error {
	return nil
}

func readonlyPath(path string) error {
	return nil
}
