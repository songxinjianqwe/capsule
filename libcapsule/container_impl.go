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
	"syscall"
	"time"
)

const (
	EnvConfigPipe      = "_LIBCAPSULE_CONFIG_PIPE"
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
Create并不会运行cmd
会让init process阻塞在cmd之前的
*/
func (c *LinuxContainerImpl) Create(process *Process) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.start(process)
}

/**
CreateAndStart
如果是exec（即不是init cmd），则在start中就会执行cmd，不需要exec再通知
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

/**
取消init process的阻塞
*/
func (c *LinuxContainerImpl) Start() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.exec()
}

func (c *LinuxContainerImpl) Destroy() error {
	panic("implement me")
}

func (c *LinuxContainerImpl) Signal(s os.Signal, all bool) error {
	panic("implement me")
}

// ************************************************************************************************
// private
// ************************************************************************************************

/**
1. parent start child
2.1 parent init, then send config
2.2 child init, then wait config
3. child wait parent config
4. parent send config
5.1 parent continue to init, then wait child SIGUSR1 signal
5.2 child continue to init, then send signal
6. child init complete, send SIGUSR1 signal to parent
7. parent received signal, then refresh state
8. child wait parent SIGUSR2 signal
9. if create, then parent exit; if run, then parent send SIGUSR2 signal to child
10. child received SIGUSR2 signal, then exec command
*/
func (c *LinuxContainerImpl) start(process *Process) error {
	logrus.Infof("LinuxContainerImpl starting...")
	// 容器启动会涉及两个管道，一个是用来传输配置信息的，一个是用来控制exec是否执行的
	// 1、创建parent process
	parent, err := NewParentProcess(c, process)
	if err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "creating new parent process")
	}
	logrus.Infof("new parent process complete, parent process: %#v", parent)
	c.initProcess = parent
	// 2、启动parent process,直至child表示自己初始化完毕，等待执行命令
	if err := parent.start(); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "starting container process")
	}
	if process.Init {
		// 3、更新容器状态
		c.createdTime = time.Now().UTC()
		c.containerState = &CreatedState{
			c: c,
		}
		// 4、持久化容器状态
		if err = c.saveState(); err != nil {
			return err
		}
	}
	return nil
}

// 让init process开始执行真正的cmd
func (c *LinuxContainerImpl) exec() error {
	logrus.Infof("send SIGUSR2 to child process...")
	return c.initProcess.signal(syscall.SIGUSR2)
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
		Status:               c.containerState.status().String(),
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
func (c *LinuxContainerImpl) saveState() error {
	state, err := c.currentState()
	if err != nil {
		return err
	}
	logrus.Infof("current state is %#v", state)
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
	if _, err = file.WriteString(string(bytes)); err != nil {
		return err
	}
	logrus.Infof("save state complete")
	return nil
}
