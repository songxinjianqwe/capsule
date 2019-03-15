package libcapsule

import (
	"encoding/json"
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/cgroups"
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
	"github.com/songxinjianqwe/capsule/libcapsule/constant"
	"github.com/songxinjianqwe/capsule/libcapsule/network"
	"github.com/songxinjianqwe/capsule/libcapsule/util"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
	"github.com/songxinjianqwe/capsule/libcapsule/util/proc"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
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
	util.PrintSubsystemPids("memory", c.id, "after signal child SIGUSR2", false)
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
		Endpoint:             *c.endpoint,
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

/*
可以根据容器进程状态判断:
1. 如果进程不存在，或状态异常，则说明为Stopped
2. 如果进程存在，那么有可能是Created或Running，从进程状态没有办法区别
3. parent process在创建容器之后会创建一个标记文件，标记容器尚未执行init process命令
4. parent process在启动容器之后会删除该文件。
*/
func (c *LinuxContainer) detectContainerStatus() (ContainerStatus, error) {
	if c.parentProcess == nil {
		return Stopped, nil
	}
	pid := c.parentProcess.pid()
	processState, err := proc.GetProcessStat(pid)
	if err != nil {
		return Stopped, nil
	}
	initProcessStartTime, _ := c.parentProcess.startTime()
	if processState.StartTime != initProcessStartTime || processState.Status == proc.Zombie || processState.Status == proc.Dead {
		return Stopped, nil
	}
	// 容器进程存在的话，会有两种情况：一种是调用完create方法，容器进程阻塞在cmd之前；一种是容器进程解除阻塞，执行了cmd
	// 在容器创建后，会创建该标记；在容器启动后，会删除该标记
	// 如果标记存在，则说明是创建容器之后，启动容器之前
	if _, err := os.Stat(filepath.Join(c.root, constant.NotExecFlagFilename)); err == nil {
		return Created, nil
	}
	return Running, nil
}

/*
检测并刷新状态，调用完该方法后，容器的containerStatusBehavior是最新的状态对象
*/
func (c *LinuxContainer) refreshStatus() error {
	detectedStatus, err := c.detectContainerStatus()
	if err != nil {
		return err
	}
	if c.statusBehavior.status() != detectedStatus {
		containerState, err := NewContainerStatusBehavior(detectedStatus, c)
		if err != nil {
			return err
		}
		if err := c.statusBehavior.transition(containerState); err != nil {
			return err
		}
	}
	return nil
}

/*
更新容器状态文件state.json
这个文件中不存储真正容器的状态，只需要在创建容器后创建文件即可，此后不再修改
*/
func (c *LinuxContainer) saveState() error {
	state, err := c.currentState()
	if err != nil {
		return err
	}
	logrus.Infof("current state is %#v", state)
	stateFilePath := filepath.Join(c.root, constant.StateFilename)
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

func (c *LinuxContainer) createFlagFile() error {
	flagFilePath := filepath.Join(c.root, constant.NotExecFlagFilename)
	logrus.Infof("creating not exec flag in file: %s", flagFilePath)
	file, err := os.Create(flagFilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	logrus.Infof("save flag complete")
	return nil
}

func (c *LinuxContainer) deleteFlagFileIfExists() error {
	flagFilePath := filepath.Join(c.root, constant.NotExecFlagFilename)
	_, err := os.Stat(flagFilePath)
	if err == nil {
		// 如果文件存在，则删除
		logrus.Infof("deleting flag :%s", flagFilePath)
		return os.Remove(flagFilePath)
	}
	return nil
}

// ****************************************************************************************************
// util
// ****************************************************************************************************
// 如果init process已经停止，则忽略terminate异常
func (c *LinuxContainer) ignoreTerminateErrors(err error) error {
	if err == nil {
		return nil
	}
	s := err.Error()
	switch {
	case strings.Contains(s, "process already finished"), strings.Contains(s, "Wait was already called"):
		return nil
	}
	return err
}

/*
构造一个command对象
*/
func (c *LinuxContainer) buildCommand(process *Process, childConfigPipe *os.File) (*exec.Cmd, error) {
	cmd := exec.Command(constant.ContainerInitCmd, constant.ContainerInitArgs)
	// 注意！Exec进程不需要新建namespace，而是进入已有的namespace
	if process.Init {
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Cloneflags: c.config.Namespaces.CloneFlags(),
		}
	}
	cmd.Dir = c.config.Rootfs
	cmd.ExtraFiles = append(cmd.ExtraFiles, childConfigPipe)
	cmd.Env = append(cmd.Env,
		fmt.Sprintf(constant.EnvConfigPipe+"=%d", constant.DefaultStdFdCount+len(cmd.ExtraFiles)-1),
	)
	// 如果后台运行，则将stdout输出到日志文件中
	if process.Detach {
		var logFileName string
		// 输出重定向
		if process.Init {
			logFileName = constant.ContainerInitLogFilename
		} else {
			logFileName = fmt.Sprintf(constant.ContainerExecLogFilenamePattern, process.ID)
		}
		logFile, err := os.Create(path.Join(constant.ContainerRuntimeRoot, c.id, logFileName))
		if err != nil {
			return nil, err
		}
		cmd.Stdout = logFile
	} else {
		// 如果启用终端，则将进程的stdin等置为os的
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd, nil
}

/*
创建一个ParentProcess实例，用于启动容器进程
有可能是InitParentProcess，也有可能是ExecParentProcess
*/
func (c *LinuxContainer) newParentProcess(process *Process) (ParentProcess, error) {
	logrus.Infof("new parent process...")
	logrus.Infof("creating pipes...")
	// socket pair 双方都可以既写又读,而pipe只能一个写,一个读
	parentConfigPipe, childConfigPipe, err := util.NewSocketPair("init")
	if err != nil {
		return nil, err
	}
	logrus.Infof("create config pipe complete, childConfigPipe: %#v, configPipe: %#v", childConfigPipe, parentConfigPipe)
	cmd, err := c.buildCommand(process, parentConfigPipe)
	if err != nil {
		return nil, err
	}
	if process.Init {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", constant.EnvInitializerType, string(StandardInitializer)))
		logrus.Infof("build command complete, command: %#v", cmd)
		logrus.Infof("new parent init process...")
		initProcess := &ParentInitProcess{
			initProcessCmd:   cmd,
			parentConfigPipe: childConfigPipe,
			container:        c,
			process:          process,
		}
		// exec process不会赋到container.parentProcess,因为它的pid,startTime返回的都exec process的,而非nochild process
		c.parentProcess = initProcess
		return initProcess, nil
	} else {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", constant.EnvInitializerType, string(ExecInitializer)))
		logrus.Infof("build command complete, command: %#v", cmd)
		logrus.Infof("new parent exec process...")
		return &ParentExecProcess{
			execProcessCmd:   cmd,
			parentConfigPipe: childConfigPipe,
			container:        c,
			process:          process,
		}, nil
	}
}
