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
	"runtime/debug"
	"strconv"
	"strings"
)

const (
	// 容器状态文件的文件名
	StateFilename = "containerState.json"
	// 用于parent进程与init进程的start/run切换
	ExecFifoFilename = "exec.fifo"
	// 重新执行本应用的command，相当于 重新执行./rune
	ContainerInitPath = "/proc/self/exe"
	// 运行容器init进程的命令
	ContainerInitArgs = "init"
	// 运行时文件的存放目录
	RuntimeRoot = "/run/rune"
)

func NewFactory() (Factory, error) {
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
	container := &LinuxContainerImpl{
		id:            id,
		root:          factory.Root,
		config:        *config,
		cgroupManager: cgroups.NewCroupManager(config.Cgroups, nil),
	}
	container.containerState = &StoppedState{c: container}
	return container, nil
}

func (factory *LinuxContainerFactoryImpl) Load(id string) (Container, error) {
	panic("implement me")
}

func (factory *LinuxContainerFactoryImpl) StartInitialization() error {
	initPipeEnv := os.Getenv(InitPipeEnv)
	initPipeFd, err := strconv.Atoi(initPipeEnv)
	if err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "converting InitPipeEnv to int")
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
	defer func() {
		// We have an error during the initialization of the container's init,
		// send it back to the parent process in the form of an initError.
		initErr := err
		bytes, err := json.Marshal(initErr)
		_, err = childPipe.WriteString(string(bytes))
		if err != nil {
			logrus.Errorf(err.Error())
			return
		}
	}()
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic from initialization: %v, %v", e, string(debug.Stack()))
		}
	}()
	// 环境变量设置
	if err := populateProcessEnvironment(initConfig.ProcessConfig.Env); err != nil {
		return util.NewGenericErrorWithInfo(err, util.SystemError, "populating environment variables")
	}
	initializer := NewInitializer(initConfig, childPipe)
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
		if err := os.Setenv(p[0], p[1]); err != nil {
			return err
		}
	}
	return nil
}
