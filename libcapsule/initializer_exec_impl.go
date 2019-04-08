package libcapsule

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/constant"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

type InitializerExecImpl struct {
	config        *InitExecConfig
	configPipe    *os.File
	containerRoot string
}

func (initializer *InitializerExecImpl) Init() error {
	logrus.WithField("exec", true).Infof("InitializerExecImpl Init()")
	// 如果后台运行，则将stdout输出到日志文件中
	if initializer.config.ProcessConfig.Detach {
		// 输出重定向
		// /var/run/capsule/containers/$container_id/exec-$exec-id.log
		logFile, err := os.OpenFile(filepath.Join(initializer.containerRoot, fmt.Sprintf(constant.ContainerExecLogFilenamePattern, initializer.config.ProcessConfig.ID)), os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0644)
		if err != nil {
			return err
		}
		if err := syscall.Dup2(int(logFile.Fd()), 1); err != nil {
			return err
		}
		if err := syscall.Dup2(int(logFile.Fd()), 2); err != nil {
			return err
		}
	}
	// look path 可以在系统的PATH里面寻找命令的绝对路径
	name, err := exec.LookPath(initializer.config.ProcessConfig.Args[0])
	if err != nil {
		return exception.NewGenericErrorWithContext(err, exception.LookPathError, "exec process/look path cmd")
	}
	logrus.WithField("exec", true).Infof("look path: %s", name)
	// syscall.Exec与cmd.Start不同，后者是启动一个新的进程来执行命令
	// 而前者会在覆盖当前进程的镜像、数据、堆栈等信息，包括PID。
	logrus.WithField("exec", true).Infof("syscall.Exec(name: %s, args: %v, env: %v)...", name, initializer.config.ProcessConfig.Args, os.Environ())
	if err := syscall.Exec(name, initializer.config.ProcessConfig.Args, os.Environ()); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SyscallExecuteCmdError, "start exec process")
	}
	return nil
}
