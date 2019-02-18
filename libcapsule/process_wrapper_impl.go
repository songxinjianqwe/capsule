package libcapsule

import (
	"encoding/json"
	"fmt"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"github.com/songxinjianqwe/rune/libcapsule/util"
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
func NewParentProcess(container *LinuxContainerImpl, process *Process) (ProcessWrapper, error) {
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
	parentProcess := InitProcessWrapperImpl{
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
ProcessWrapper接口的实现类，包裹了InitProcess，它返回的进程信息均为容器Init进程的信息
*/
type InitProcessWrapperImpl struct {
	initProcessCmd *exec.Cmd
	parentPipe     *os.File
	childPipe      *os.File
	container      *LinuxContainerImpl
	process        *Process
	bootstrapData  io.Reader
}

type InitConfig struct {
	ContainerConfig configc.Config
	ProcessConfig   Process
}

func (p *InitProcessWrapperImpl) pid() int {
	panic("implement me")
}

func (p *InitProcessWrapperImpl) start() error {
	defer p.parentPipe.Close()
	err := p.initProcessCmd.Start()
	if err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "starting init process command")
	}
	if err = p.createNetworkInterfaces(); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "creating network interfaces")
	}
	if err = p.sendConfig(); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "sending config to init process")
	}
	if err = p.parentPipe.Close(); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "closing parent pipe")
	}
	p.wait()
	return nil
}

func (p *InitProcessWrapperImpl) terminate() error {
	panic("implement me")
}

func (p *InitProcessWrapperImpl) wait() (*os.ProcessState, error) {
	panic("implement me")
}

func (p *InitProcessWrapperImpl) startTime() (uint64, error) {
	panic("implement me")
}

func (p *InitProcessWrapperImpl) signal(os.Signal) error {
	panic("implement me")
}

func (p *InitProcessWrapperImpl) externalDescriptors() []string {
	panic("implement me")
}

func (p *InitProcessWrapperImpl) setExternalDescriptors(fds []string) {
	panic("implement me")
}

func (p *InitProcessWrapperImpl) createNetworkInterfaces() error {
	panic("implement me")
}

func (p *InitProcessWrapperImpl) sendConfig() error {
	initConfig := &InitConfig{
		ContainerConfig: p.container.config,
		ProcessConfig:   *p.process,
	}
	bytes, err := json.Marshal(initConfig)
	if err != nil {
		return err
	}
	_, err = p.parentPipe.WriteString(string(bytes))
	return err
}
