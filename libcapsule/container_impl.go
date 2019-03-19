package libcapsule

import (
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/cgroups"
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
	"github.com/songxinjianqwe/capsule/libcapsule/network"
	"github.com/songxinjianqwe/capsule/libcapsule/util"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
	"os"
	"sync"
	"syscall"
	"time"
)

type LinuxContainer struct {
	id     string
	root   string
	config configs.ContainerConfig
	// runtime info
	cgroupManager  cgroups.CgroupManager
	endpoint       *network.Endpoint
	parentProcess  ParentProcess
	statusBehavior ContainerStatusBehavior
	createdTime    time.Time
	mutex          sync.Mutex
}

// ************************************************************************************************
// public
// ************************************************************************************************

func (c *LinuxContainer) ID() string {
	return c.id
}

func (c *LinuxContainer) Status() (ContainerStatus, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.currentStatus()
}

func (c *LinuxContainer) State() (*StateStorage, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.currentState()
}

func (c *LinuxContainer) OCIState() (*specs.State, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.currentOCIState()
}

func (c *LinuxContainer) Config() configs.ContainerConfig {
	return c.config
}

/*
Create并不会运行cmd
会让init process阻塞在cmd之前的
*/
func (c *LinuxContainer) Create(process *Process) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.create(process)
}

/*
CreateAndStart
如果是exec（即不是init cmd），则在start中就会执行cmd，不需要exec再通知
*/
func (c *LinuxContainer) Run(process *Process) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if err := c.create(process); err != nil {
		return err
	}
	if process.Init {
		if err := c.start(); err != nil {
			return err
		}
	}
	return nil
}

/*
取消init process的阻塞
*/
func (c *LinuxContainer) Start() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.start()
}

func (c *LinuxContainer) Destroy() error {
	c.mutex.Lock()
	c.mutex.Unlock()
	return c.statusBehavior.destroy()
}

func (c *LinuxContainer) Signal(s os.Signal) error {
	c.mutex.Lock()
	c.mutex.Unlock()
	status, err := c.currentStatus()
	logrus.Infof("sending %s to container init process...", s)
	if err != nil {
		return err
	}
	if status == Running || status == Created {
		if err := c.parentProcess.signal(s); err != nil {
			return exception.NewGenericErrorWithContext(err, exception.ParentProcessSignalError, "signaling init process")
		}
		return nil
	}
	return exception.NewGenericErrorWithContext(err, exception.ContainerNotRunningError, "signaling init process")
}

// ************************************************************************************************
// private
// ************************************************************************************************

/*
1. parent create child
2.1 parent init, then send config
2.2 child init, then wait config
3. child wait parent config
4. parent send config
5.1 parent continue to init, then wait child SIGUSR1/SIGCHLD signal
5.2 child continue to init, then send signal
6. child init complete/failed, send SIGUSR1/SIGCHLD signal to parent
7. parent received signal, then refresh state
8. child wait parent SIGUSR2 signal
9. if create, then parent exit; if run, then parent send SIGUSR2 signal to child
10. child received SIGUSR2 signal, then start command
*/
func (c *LinuxContainer) create(process *Process) error {
	logrus.Infof("LinuxContainer starting...")
	// 1、创建parent config
	parent, err := c.newParentProcess(process)
	if err != nil {
		return exception.NewGenericErrorWithContext(err, exception.ParentProcessCreateError, "creating new parent process")
	}
	logrus.Infof("new parent process complete, parent config: %#v", parent)
	// 2、启动parent config,直至child表示自己初始化完毕，等待执行命令
	if err := parent.start(); err != nil {
		// 启动失败，则杀掉init process，如果是已经停止，则忽略。
		logrus.Warnf("parent process init/exec failed, killing init/exec process...")
		if err := c.ignoreTerminateErrors(parent.terminate()); err != nil {
			logrus.Warn(err)
		}
		return exception.NewGenericErrorWithContext(err, exception.ParentProcessStartError, "starting container process")
	}
	if process.Init {
		// 3、更新容器状态
		c.createdTime = time.Now()
		c.statusBehavior = &CreatedStatusBehavior{
			c: c,
		}
		// 4、持久化容器状态
		if err = c.saveState(); err != nil {
			return err
		}
		// 5、创建标记文件，表示Created
		if err := c.createFlagFile(); err != nil {
			return err
		}
	}
	logrus.Infof("create/exec container complete!")
	return nil
}

// 让init process开始执行真正的cmd
func (c *LinuxContainer) start() error {
	logrus.Infof("container starting...")
	// 目前一定是Created状态
	util.PrintSubsystemPids("memory", c.id, "before container start", false)

	logrus.Infof("send SIGUSR2 to child process...")
	if err := c.parentProcess.signal(syscall.SIGUSR2); err != nil {
		return err
	}
	// 这里不好判断是否是之前在运行的是否是init process，索性就 有就删，没有就算了
	if err := c.deleteFlagFileIfExists(); err != nil {
		return err
	}
	logrus.Infof("refreshing container status...")
	if err := c.refreshStatus(); err != nil {
		return err
	}
	// 对于前台进程来说，这里必须wait，否则在仅有容器进程存活情况下，它在输入任何命令后立即退出，并且ssh进程退出/登录用户注销
	if !c.parentProcess.detach() {
		logrus.Infof("wait child process exit...")
		if err := c.parentProcess.wait(); err != nil {
			return exception.NewGenericErrorWithContext(err, exception.ParentProcessWaitError, "waiting child process exit")
		}
		logrus.Infof("child process exited")
	}
	return nil
}

func (c *LinuxContainer) currentState() (*StateStorage, error) {
	initProcessPid := c.parentProcess.pid()
	initProcessStartTime, _ := c.parentProcess.startTime()
	state := &StateStorage{
		ID:                   c.ID(),
		Config:               c.config,
		InitProcessPid:       initProcessPid,
		InitProcessStartTime: initProcessStartTime,
		Created:              c.createdTime,
		CgroupPaths:          c.cgroupManager.GetPaths(),
		NamespacePaths:       make(map[configs.NamespaceType]string),
		Endpoint:             c.endpoint,
	}
	if initProcessPid > 0 {
		for _, ns := range c.config.Namespaces {
			state.NamespacePaths[ns.Type] = ns.GetPath(initProcessPid)
		}
	}
	return state, nil
}

func (c *LinuxContainer) currentOCIState() (*specs.State, error) {
	bundle, annotations := util.GetAnnotations(c.config.Labels)
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
		if c.parentProcess != nil {
			state.Pid = c.parentProcess.pid()
		}
	}
	return state, err
}

/*
容器状态可以存储在state.json文件中，也可以每次去检测。
前者是不靠谱的！如果是后台运行的容器，那么在parent process结束后，容器可能会退出，但此时parent process不会
去监听容器进程状态，也就无法保证state.json文件的状态总是正确的。
后者是每次获取状态时都去检测一遍，并矫正内存状态。
*/
func (c *LinuxContainer) currentStatus() (ContainerStatus, error) {
	if err := c.refreshStatus(); err != nil {
		return -1, err
	}
	return c.statusBehavior.status(), nil
}
