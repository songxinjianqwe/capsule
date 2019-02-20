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
	defer func() {
		if e := recover(); e != nil {
			logrus.Errorf("panic from initialization: %v, %v", e, string(debug.Stack()))
		}
	}()
	configPipeEnv := os.Getenv(EnvConfigPipe)
	initPipeFd, err := strconv.Atoi(configPipeEnv)
	logrus.WithField("init", true).Infof("got config pipe env: %d", initPipeFd)
	if err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "converting EnvConfigPipe to int")
	}
	initializerType := InitializerType(os.Getenv(EnvInitializerType))
	logrus.WithField("init", true).Infof("got initializer type: %s", initializerType)

	execPipeFd := -1
	// 只有init process有fifo管道
	if initializerType == StandardInitializer {
		execPipeEnv := os.Getenv(EnvExecPipe)
		execPipeFd, err = strconv.Atoi(execPipeEnv)
		logrus.WithField("init", true).Infof("got exec pipe env: %d", execPipeFd)
		if err != nil {
			return util.NewGenericErrorWithInfo(err, util.SystemError, "converting EnvConfigPipe to int")
		}
	}

	configPipe := os.NewFile(uintptr(initPipeFd), "pipe")
	logrus.WithField("init", true).Infof("open child pipe: %#v", configPipe)
	logrus.WithField("init", true).Infof("starting to read init config from child pipe")
	bytes, err := ioutil.ReadAll(configPipe)
	if err != nil {
		logrus.WithField("init", true).Errorf("read init config failed: %s", err.Error())
		return util.NewGenericErrorWithInfo(err, util.SystemError, "reading init config from configPipe")
	}
	// child 读完就关
	if err = configPipe.Close(); err != nil {
		logrus.Errorf("closing parent pipe failed: %s", err.Error())
	}
	logrus.Infof("read init config complete, unmarshal bytes")
	initConfig := &InitConfig{}
	if err = json.Unmarshal(bytes, initConfig); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "unmarshal init config")
	}
	logrus.WithField("init", true).Infof("read init config from child pipe: %#v", initConfig)

	// 环境变量设置
	if err := populateProcessEnvironment(initConfig.ProcessConfig.Env); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "populating environment variables")
	}
	initializer, err := NewInitializer(initializerType, initConfig, configPipe, execPipeFd)
	if err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "creating initializer")
	}
	logrus.WithField("init", true).Infof("created initializer:%#v", initializer)
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
		logrus.WithField("init", true).Infof("set env: key:%s, value:%s", p[0], p[1])
		if err := os.Setenv(p[0], p[1]); err != nil {
			return err
		}
	}
	return nil
}
