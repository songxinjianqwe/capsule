package libcapsule

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

/**
一个进程默认有三个文件描述符，stdin、stdout、stderr
外带的文件描述符在这三个fd之后
*/
const DefaultStdFdCount = 3

/**
创建一个ProcessWrapper实例，用于启动容器进程
有可能是InitProcessWrapper，也有可能是SetnsProcessWrapper
*/
func NewParentProcess(container *LinuxContainerImpl, process *Process) (ProcessWrapper, error) {
	logrus.Infof("new parent process...")
	logrus.Infof("creating pipe...")
	reader, writer, err := os.Pipe()
	logrus.Infof("create pipe complete, parent: %#v, child: %#v", writer, reader)
	if err != nil {
		return nil, err
	}
	cmd, err := buildCommand(container,
		process, reader, process.Init)
	logrus.Infof("build command complete, command: %#v", cmd)
	if err != nil {
		return nil, err
	}
	if process.Init {
		return NewInitProcessWrapper(process, cmd, writer, reader, container), nil
	} else {
		return NewSetnsProcessWrapper(process, cmd, writer, reader), nil
	}
}

/**
构造一个command对象
*/
func buildCommand(container *LinuxContainerImpl, process *Process, childPipe *os.File, init bool) (*exec.Cmd, error) {
	cmd := exec.Command(ContainerInitPath, ContainerInitArgs)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: container.config.Namespaces.CloneFlags(),
	}
	cmd.Dir = container.config.Rootfs
	cmd.ExtraFiles = append(cmd.ExtraFiles, childPipe)
	cmd.Env = append(cmd.Env,
		fmt.Sprintf(EnvInitPipe+"=%d", DefaultStdFdCount+len(cmd.ExtraFiles)-1),
	)
	if init {
		fifo, err := os.OpenFile(filepath.Join(container.root, ExecFifoFilename), os.O_RDWR, 0)
		if err != nil {
			return nil, err
		}
		cmd.ExtraFiles = append(cmd.ExtraFiles, fifo)
		cmd.Env = append(cmd.Env,
			fmt.Sprintf(EnvExecFifo+"=%d", DefaultStdFdCount+len(cmd.ExtraFiles)-1),
		)
	}
	cmd.Stdin = process.Stdin
	cmd.Stdout = process.Stdout
	cmd.Stderr = process.Stderr
	return cmd, nil
}
