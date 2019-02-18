package libcapsule

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

/**
一个进程默认有三个文件描述符，stdin、stdout、stderr
外带的文件描述符在这三个fd之后
*/
const DefaultStdFdCount = 3

/**
创建一个ParentProcess实例，用于启动容器Init进程，并和容器Init进程通信
*/
func NewParentProcess(container *LinuxContainer, process *Process) (ParentProcess, error) {
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	initProcessCmd, err := buildInitProcessCommand(container.config.Rootfs,
		container.config.Namespaces.CloneFlags(),
		process, reader)
	if err != nil {
		return nil, err
	}
	parentProcess := InitProcessRunner{
		initProcessCmd: initProcessCmd,
		parentPipe:     writer,
		childPipe:      reader,
		container:      container,
		process:        process,
	}
	return &parentProcess, nil
}

/**
构造一个init进程的command对象
*/
func buildInitProcessCommand(cmdDir string, cloneFlags uintptr, process *Process, childPipe *os.File) (*exec.Cmd, error) {
	cmd := exec.Command(ContainerInitPath, ContainerInitArgs)
	cmd.Stdin = process.Stdin
	cmd.Stdout = process.Stdout
	cmd.Stderr = process.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: cloneFlags,
	}
	cmd.Dir = cmdDir
	cmd.ExtraFiles = append(cmd.ExtraFiles, childPipe)
	cmd.Env = append(cmd.Env,
		fmt.Sprintf(InitPipeEnv+"=%d", DefaultStdFdCount+len(cmd.ExtraFiles)-1),
	)
	return cmd, nil
}

/**
ParentProcess接口的实现类，包裹了InitProcess，它返回的进程信息均为容器Init进程的信息
*/
type InitProcessRunner struct {
	initProcessCmd *exec.Cmd
	parentPipe     *os.File
	childPipe      *os.File
	container      *LinuxContainer
	process        *Process
	bootstrapData  io.Reader
}

func (p *InitProcessRunner) pid() int {
	panic("implement me")
}

func (p *InitProcessRunner) start() error {
	panic("implement me")
}

func (p *InitProcessRunner) terminate() error {
	panic("implement me")
}

func (p *InitProcessRunner) wait() (*os.ProcessState, error) {
	panic("implement me")
}

func (p *InitProcessRunner) startTime() (uint64, error) {
	panic("implement me")
}

func (p *InitProcessRunner) signal(os.Signal) error {
	panic("implement me")
}

func (p *InitProcessRunner) externalDescriptors() []string {
	panic("implement me")
}

func (p *InitProcessRunner) setExternalDescriptors(fds []string) {
	panic("implement me")
}
