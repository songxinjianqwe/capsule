package libcapsule

import (
	"encoding/json"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/cgroups"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	EnvConfigPipe      = "_LIBCAPSULE_CONFIG_PIPE"
	EnvExecPipe        = "_LIBCAPSULE_EXEC_PIPE"
	EnvInitializerType = "_LIBCAPSULE_INITIALIZER_TYPE"
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
	logrus.Infof("LinuxContainerImpl starting...")
	// 容器启动会涉及两个管道，一个是用来传输配置信息的，一个是用来控制exec是否执行的
	// 1、创建parent process
	parent, err := NewParentProcess(c, process)
	if err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "creating new parent process")
	}
	logrus.Infof("new parent process complete, parent process: %#v", parent)
	c.initProcess = parent
	// 2、启动parent process
	if err := parent.start(); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "starting container process")
	}
	if process.Init {
		// 3、更新容器状态
		c.createdTime = time.Now().UTC()
		c.containerState = &CreatedState{
			c: c,
		}
		// 4、持久化容器状态
		_, err = c.updateState()
		if err != nil {
			return err
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
	return c.containerState.status(), nil
}

/**
更新容器状态文件state.json
*/
func (c *LinuxContainerImpl) updateState() (state *State, err error) {
	state, err = c.currentState()
	if err != nil {
		return nil, err
	}
	logrus.Infof("current state is %#v", state)
	err = c.saveState(state)
	if err != nil {
		return nil, err
	}
	logrus.Infof("save state complete")
	return state, nil
}

/**
将state JSON对象写入到文件中
*/
func (c *LinuxContainerImpl) saveState(state *State) error {
	stateFilePath := filepath.Join(c.root, StateFilename)
	logrus.Infof("saving state in file: %s", stateFilePath)
	file, err := os.Create(stateFilePath)
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
