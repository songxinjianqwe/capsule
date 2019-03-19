package libcapsule

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
	"github.com/songxinjianqwe/capsule/libcapsule/constant"
	"github.com/songxinjianqwe/capsule/libcapsule/util"
	"github.com/songxinjianqwe/capsule/libcapsule/util/proc"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

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

/*
构造一个command对象
*/
func (c *LinuxContainer) buildCommand(process *Process, childConfigPipe *os.File) (*exec.Cmd, error) {
	cmd := exec.Command(constant.ContainerInitCmd, constant.ContainerInitArgs)
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
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", constant.EnvInitializerType, string(InitInitializer)))
		logrus.Infof("build command complete, command: %#v", cmd)
		logrus.Infof("new parent init process...")
		namepaces := make(map[configs.NamespaceType]string)
		for _, ns := range c.config.Namespaces {
			if ns.Path != "" {
				namepaces[ns.Type] = ns.Path
			}
		}
		initProcess := &ParentAbstractProcess{
			processCmd:       cmd,
			parentConfigPipe: childConfigPipe,
			container:        c,
			process:          process,
			cloneFlags:       c.config.Namespaces.CloneFlagsOfEmptyPath(),
			namespacePathMap: namepaces,
			stackHook:        initStartHook,
		}
		// exec process不会赋到container.parentProcess,因为它的pid,startTime返回的都是exec process的,而非nochild process(反映的是init process的)
		c.parentProcess = initProcess
		return initProcess, nil
	} else {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", constant.EnvInitializerType, string(ExecInitializer)))
		logrus.Infof("build command complete, command: %#v", cmd)
		logrus.Infof("new parent exec process...")
		currentState, err := c.currentState()
		if err != nil {
			return nil, err
		}
		return &ParentAbstractProcess{
			processCmd:       cmd,
			parentConfigPipe: childConfigPipe,
			container:        c,
			process:          process,
			cloneFlags:       0,
			namespacePathMap: currentState.NamespacePaths,
			stackHook:        execStartHook,
		}, nil
	}
}

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
