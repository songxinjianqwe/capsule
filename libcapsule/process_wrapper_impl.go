package libcapsule

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/cgroups"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"github.com/songxinjianqwe/rune/libcapsule/util/system"
	"golang.org/x/sys/unix"
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
创建一个ParentProcess实例，用于启动容器Init进程，并和容器Init进程通信
*/
func NewParentProcess(container *LinuxContainerImpl, process *Process) (ProcessWrapper, error) {
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	fifo, err := os.OpenFile(filepath.Join(container.root, ExecFifoFilename), os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	initProcessCmd, err := buildInitProcessCommand(container.config.Rootfs,
		container.config.Namespaces.CloneFlags(),
		process, reader, fifo)
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
func buildInitProcessCommand(rootfs string, cloneFlags uintptr, process *Process, childPipe *os.File, fifo *os.File) (*exec.Cmd, error) {
	cmd := exec.Command(ContainerInitPath, ContainerInitArgs)
	cmd.Stdin = process.Stdin
	cmd.Stdout = process.Stdout
	cmd.Stderr = process.Stderr

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: cloneFlags,
	}
	cmd.Dir = rootfs
	cmd.ExtraFiles = append(cmd.ExtraFiles, childPipe)
	cmd.Env = append(cmd.Env,
		fmt.Sprintf(EnvInitPipe+"=%d", DefaultStdFdCount+len(cmd.ExtraFiles)-1),
	)
	cmd.ExtraFiles = append(cmd.ExtraFiles, fifo)
	cmd.Env = append(cmd.Env,
		fmt.Sprintf(EnvExecFifo+"=%d", DefaultStdFdCount+len(cmd.ExtraFiles)-1),
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
	cgroupManger   *cgroups.CgroupManager
}

type InitConfig struct {
	ContainerConfig configc.Config
	ProcessConfig   Process
}

func (p *InitProcessWrapperImpl) start() error {
	defer p.parentPipe.Close()
	// 非阻塞
	err := p.initProcessCmd.Start()
	p.childPipe.Close()
	if err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "starting init process command")
	}
	if err := p.container.cgroupManager.Apply(p.pid()); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "applying cgroup configuration for process")
	}
	defer func() {
		if err != nil {
			p.container.cgroupManager.Destroy()
		}
	}()
	childPid, err := p.getChildPid()
	if err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "getting the final child's pid from pipe")
	}
	if err := p.container.cgroupManager.Apply(childPid); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "applying cgroup configuration for process")
	}
	if err := p.waitForChildExit(childPid); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "waiting for our first child to exit")
	}
	if err = p.createNetworkInterfaces(); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "creating network interfaces")
	}
	// init process会在启动后阻塞，直至收到config
	if err = p.sendConfig(); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "sending config to init process")
	}
	if err = p.parentPipe.Close(); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "closing parent pipe")
	}
	state, err := p.wait()
	if err != nil {
		logrus.Errorf("waiting init process cmd error:%v, %s", state, err.Error())
	}
	return nil
}

func (p *InitProcessWrapperImpl) pid() int {
	return p.initProcessCmd.Process.Pid
}

func (p *InitProcessWrapperImpl) terminate() error {
	panic("implement me")
}

func (p *InitProcessWrapperImpl) wait() (*os.ProcessState, error) {
	panic("implement me")
}

func (p *InitProcessWrapperImpl) startTime() (uint64, error) {
	stat, err := system.GetProcessStat(p.pid())
	return stat.StartTime, err
}

func (p *InitProcessWrapperImpl) signal(sig os.Signal) error {
	s, ok := sig.(syscall.Signal)
	if !ok {
		return util.NewGenericError(fmt.Errorf("os: unsupported signal type:%v", sig), util.SystemError)
	}
	return unix.Kill(p.pid(), s)
}

// ******************************************************************************************************
// biz methods
// ******************************************************************************************************

func (p *InitProcessWrapperImpl) createNetworkInterfaces() error {
	return nil
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

func (p *InitProcessWrapperImpl) getChildPid() (int, error) {
	return -1, nil
}

func (p *InitProcessWrapperImpl) waitForChildExit(pid int) error {
	return nil
}
