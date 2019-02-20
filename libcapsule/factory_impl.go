package libcapsule

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/cgroups"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"github.com/songxinjianqwe/rune/libcapsule/configc/validate"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
)

const (
	// 容器状态文件的文件名
	// 存放在 RuntimeRoot/containerId/下
	StateFilename = "state.json"
	// 用于parent进程与init进程的start/run切换
	// 存放在 RuntimeRoot/containerId/下
	ExecFifoFilename = "exec.fifo"
	// 重新执行本应用的command，相当于 重新执行./rune
	ContainerInitPath = "/proc/self/exe"
	// 运行容器init进程的命令
	ContainerInitArgs = "init"
	// 运行时文件的存放目录
	RuntimeRoot = "/run/rune"
	// 容器配置文件，存放在运行rune的cwd下
	ContainerConfigFilename = "config.json"
)

func NewFactory() (Factory, error) {
	logrus.Infof("new container factory ...")
	logrus.Infof("mkdir RuntimeRoot: %s", RuntimeRoot)
	if err := os.MkdirAll(RuntimeRoot, 0700); err != nil {
		return nil, util.NewGenericError(err, util.SystemError)
	}
	factory := LinuxContainerFactoryImpl{
		Root:      RuntimeRoot,
		Validator: validate.New(),
	}
	return &factory, nil
}

type LinuxContainerFactoryImpl struct {
	// Root directory for the factory to store containerState.
	Root string

	// Validator provides validation to container configurations.
	Validator validate.Validator
}

func (factory *LinuxContainerFactoryImpl) Create(id string, config *configc.Config) (Container, error) {
	logrus.Infof("container factory creating container: %s", id)
	containerRoot := filepath.Join(factory.Root, id)
	if _, err := os.Stat(containerRoot); err == nil {
		return nil, util.NewGenericError(fmt.Errorf("container with id exists: %v", id), util.SystemError)
	} else if !os.IsNotExist(err) {
		return nil, util.NewGenericError(err, util.SystemError)
	}
	logrus.Infof("mkdir containerRoot: %s", containerRoot)
	if err := os.MkdirAll(containerRoot, 0711); err != nil {
		return nil, util.NewGenericError(err, util.SystemError)
	}
	container := &LinuxContainerImpl{
		id:            id,
		root:          containerRoot,
		config:        *config,
		cgroupManager: cgroups.NewCroupManager(config.Cgroups),
	}
	container.containerState = &StoppedState{c: container}
	logrus.Infof("create container complete, container: %#v", container)
	return container, nil
}

func (factory *LinuxContainerFactoryImpl) Load(id string) (Container, error) {
	panic("implement me")
}

func (factory *LinuxContainerFactoryImpl) StartInitialization() error {
	initPipeEnv := os.Getenv(EnvInitPipe)
	initPipeFd, err := strconv.Atoi(initPipeEnv)
	logrus.Infof("got init pipe env: %d", initPipeFd)
	if err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "converting EnvInitPipe to int")
	}

	initializerType := InitializerType(os.Getenv(EnvInitializerType))
	logrus.Infof("got initializer type: %s", initializerType)
	execFifoFd := -1
	// 只有init process有fifo管道
	if initializerType == StandardInitializer {
		execFifoEnv := os.Getenv(EnvExecFifo)
		execFifoFd, err = strconv.Atoi(execFifoEnv)
		logrus.Infof("got exec fifo: %d", execFifoFd)
		if err != nil {
			return util.NewGenericErrorWithInfo(err, util.SystemError, "converting EnvInitPipe to int")
		}
	}

	childPipe := os.NewFile(uintptr(initPipeFd), "pipe")
	defer childPipe.Close()
	bytes, err := ioutil.ReadAll(childPipe)
	if err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "reading init config from childPipe")
	}
	initConfig := &InitConfig{}
	if err = json.Unmarshal(bytes, initConfig); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "unmarshalling init config")
	}
	logrus.Infof("read init config from child pipe: %#v", initConfig)
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic from initialization: %v, %v", e, string(debug.Stack()))
		}
	}()
	// 环境变量设置
	if err := populateProcessEnvironment(initConfig.ProcessConfig.Env); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "populating environment variables")
	}
	initializer, err := NewInitializer(initializerType, initConfig, childPipe, execFifoFd)
	if err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "creating initializer")
	}
	logrus.Infof("created initializer:%#v", initializer)
	if err := initializer.Init(); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "executing init command")
	}
	return nil
}

// populateProcessEnvironment loads the provided environment variables into the
// current processes's environment.
func populateProcessEnvironment(env []string) error {
	for _, pair := range env {
		p := strings.SplitN(pair, "=", 2)
		if len(p) < 2 {
			return fmt.Errorf("invalid environment '%v'", pair)
		}
		logrus.Infof("set env: key:%s, value:%s", p[0], p[1])
		if err := os.Setenv(p[0], p[1]); err != nil {
			return err
		}
	}
	return nil
}
