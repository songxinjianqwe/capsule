package libcapsule

import (
	"encoding/json"
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/cgroups"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"github.com/songxinjianqwe/rune/libcapsule/util/proc"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	EnvInitPipe        = "_LIBCAPSULE_INITPIPE"
	EnvExecFifo        = "_LIBCAPSULE_EXECFIFO"
	EnvInitializerType = "_LIBCAPSULE_INITIALIZERTYPE"
)

type LinuxContainerImpl struct {
	id             string
	root           string
	config         configc.Config
	cgroupManager  cgroups.CgroupManager
	initProcess    ProcessWrapper
	containerState ContainerState
	createdTime    time.Time
	mutex          sync.Mutex
}

// ************************************************************************************************
// public
// ************************************************************************************************

func (c *LinuxContainerImpl) ID() string {
	return c.id
}

func (c *LinuxContainerImpl) Status() (Status, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.currentStatus()
}

func (c *LinuxContainerImpl) State() (*State, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.currentState()
}

func (c *LinuxContainerImpl) OCIState() (*specs.State, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.currentOCIState()
}

func (c *LinuxContainerImpl) Config() configc.Config {
	return c.config
}

func (c *LinuxContainerImpl) Processes() ([]int, error) {
	panic("implement me")
}

/**
Start是会让init process阻塞在cmd之前的
*/
func (c *LinuxContainerImpl) Start(process *Process) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.start(process)
}

/**
Run 当运行init process，不会阻塞，会执行完cmd
当运行非init process，会阻塞
*/
func (c *LinuxContainerImpl) Run(process *Process) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if err := c.start(process); err != nil {
		return err
	}
	if process.Init {
		if err := c.exec(); err != nil {
			return err
		}
	}
	return nil
}

func (c *LinuxContainerImpl) Destroy() error {
	panic("implement me")
}

func (c *LinuxContainerImpl) Signal(s os.Signal, all bool) error {
	panic("implement me")
}

/**
取消init process的阻塞
*/
func (c *LinuxContainerImpl) Exec() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.exec()
}

// ************************************************************************************************
// private
// ************************************************************************************************

func (c *LinuxContainerImpl) start(process *Process) error {
	// 容器启动会涉及两个管道，一个是用来传输配置信息的，一个是用来控制exec是否执行的
	// 1、创建exec管道文件
	if process.Init {
		if err := c.createExecFifo(); err != nil {
			return err
		}
	}
	// 2、创建parent process
	parent, err := NewParentProcess(c, process)
	if err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "creating new parent process")
	}
	c.initProcess = parent
	// 3、启动parent process
	if err := parent.start(); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "starting container process")
	}
	if process.Init {
		// 4、更新容器状态
		c.createdTime = time.Now().UTC()
		c.containerState = &CreatedState{
			c: c,
		}
		// 5、持久化容器状态
		_, err = c.updateState()
		if err != nil {
			return err
		}
		// 6、删除exec管道文件
		err = c.deleteExecFifo()
		if err != nil {
			logrus.Errorf("delete exec fifo failed: %s", err.Error())
		}
	}
	return nil
}

func (c *LinuxContainerImpl) exec() error {
	panic("implement me")
}

func (c *LinuxContainerImpl) currentState() (*State, error) {
	var (
		initProcessPid       = -1
		initProcessStartTime uint64
	)
	if c.initProcess != nil {
		initProcessPid = c.initProcess.pid()
		initProcessStartTime, _ = c.initProcess.startTime()
	}
	state := &State{
		ID:                   c.ID(),
		Config:               c.config,
		InitProcessPid:       initProcessPid,
		InitProcessStartTime: initProcessStartTime,
		Created:              c.createdTime,
		CgroupPaths:          c.cgroupManager.GetPaths(),
		NamespacePaths:       make(map[configc.NamespaceType]string),
	}
	if initProcessPid > 0 {
		for _, ns := range c.config.Namespaces {
			state.NamespacePaths[ns.Type] = ns.GetPath(initProcessPid)
		}
	}
	return state, nil
}

func (c *LinuxContainerImpl) currentOCIState() (*specs.State, error) {
	bundle, annotations := util.Annotations(c.config.Labels)
	state := &specs.State{
		Version:     specs.Version,
		ID:          c.ID(),
		Bundle:      bundle,
		Annotations: annotations,
	}
	status, err := c.currentStatus()
	if err != nil {
		return nil, err
	}
	state.Status = status.String()
	if status != Stopped {
		if c.initProcess != nil {
			state.Pid = c.initProcess.pid()
		}
	}
	return state, err
}

func (c *LinuxContainerImpl) currentStatus() (Status, error) {
	if err := c.refreshState(); err != nil {
		return -1, err
	}
	return c.containerState.status(), nil
}

/**
因为外部用户来使用libcapsule时会随意修改容器内存中的状态，所以这个containerState并不可信，我们需要自己去检测一个真正的容器状态，
并进行数据纠错
*/
func (c *LinuxContainerImpl) refreshState() error {
	status, err := c.detectRealStatus()
	if err != nil {
		return err
	}
	switch status {
	case Created:
		return c.containerState.transition(&CreatedState{c: c})
	case Running:
		return c.containerState.transition(&RunningState{c: c})
	case Stopped:
		return c.containerState.transition(&StoppedState{c: c})
	default:
		return util.NewGenericError(fmt.Errorf("检测到未知容器状态:%d", status), util.SystemError)
	}
}

/**
1、如果容器init进程不存在，或者进程已经死亡或成为僵尸进程，则均为 【Stopped】
2、如果exec.fifo文件存在，则为 【Created】
3、其他情况为 【Running】
*/
func (c *LinuxContainerImpl) detectRealStatus() (Status, error) {
	if c.initProcess == nil {
		return Stopped, nil
	}
	pid := c.initProcess.pid()
	processState, err := proc.Stat(pid)
	if err != nil {
		return Stopped, nil
	}
	initProcessStartTime, _ := c.initProcess.startTime()
	if processState.StartTime != initProcessStartTime || processState.State == proc.Zombie || processState.State == proc.Dead {
		return Stopped, nil
	}
	// 在容器创建前，会先创建exec管道；在容器创建后，会删除该管道
	if _, err := os.Stat(filepath.Join(c.root, ExecFifoFilename)); err == nil {
		return Created, nil
	}
	return Running, nil
}

/**
在start前，创建exec.fifo管道
io.Pipe是内存管道，无法通过内存管道来感知容器状态
因为管道存在，则说明容器是处于created之后，running之前的状态
*/
func (c *LinuxContainerImpl) createExecFifo() error {
	fifoName := filepath.Join(c.root, ExecFifoFilename)

	if _, err := os.Stat(fifoName); err == nil {
		return fmt.Errorf("exec fifo %s already exists", fifoName)
	}
	// 读是4，写是2，执行是1
	// 自己可以读写，同组可以写，其他组可以写
	if err := unix.Mkfifo(fifoName, 0622); err != nil {
		return err
	}
	return nil
}

/**
在start后，删除exec.fifo管道
*/
func (c *LinuxContainerImpl) deleteExecFifo() error {
	fifoName := filepath.Join(c.root, ExecFifoFilename)
	return os.Remove(fifoName)
}

/**
更新容器状态文件state.json
*/
func (c *LinuxContainerImpl) updateState() (state *State, err error) {
	state, err = c.currentState()
	if err != nil {
		return nil, err
	}
	err = c.saveState(state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

/**
将state JSON对象写入到文件中
*/
func (c *LinuxContainerImpl) saveState(state *State) error {
	file, err := os.Create(filepath.Join(RuntimeRoot, StateFilename))
	if err != nil {
		return err
	}
	defer file.Close()
	bytes, err := json.Marshal(state)
	if err != nil {
		return err
	}
	_, err = file.WriteString(string(bytes))
	return err
}
