package libcapsule

import (
	"github.com/songxinjianqwe/rune/cli/constant"
	"github.com/songxinjianqwe/rune/libcapsule/cgroups"
	"github.com/songxinjianqwe/rune/libcapsule/config"
	"github.com/songxinjianqwe/rune/libcapsule/config/validate"
	"os"
)

const (
	stateFilename    = "state.json"
	execFifoFilename = "exec.fifo"
)

func NewFactory() (Factory, error) {
	factory := LinuxContainerFactory{
		Root:     constant.RuntimeRoot,
		InitPath: "/proc/self/exe",
		// 第一个元素是真正命令，比如create,run等；第二个元素是用来调用init进程的
		InitArgs:  []string{os.Args[0], "init"},
		Validator: validate.New(),
		// TODO
		NewCgroupsManager: nil,
	}
	return &factory, nil
}

type LinuxContainerFactory struct {
	// Root directory for the factory to store state.
	Root string

	// InitPath is the path for calling the init responsibilities for spawning
	// a container.
	InitPath string

	// InitArgs are arguments for calling the init responsibilities for spawning
	// a container.
	InitArgs []string

	// Validator provides validation to container configurations.
	Validator validate.Validator

	// NewCgroupsManager returns an initialized cgroups manager for a single container.
	NewCgroupsManager func(config *config.Cgroup, paths map[string]string) cgroups.CgroupManager
}

func (factory *LinuxContainerFactory) Create(id string, config *config.Config) (Container, error) {
	container := LinuxContainer{
		id:            id,
		root:          factory.Root,
		config:        *config,
		cgroupManager: factory.NewCgroupsManager(config.Cgroups, nil),
		initPath:      factory.InitPath,
		initArgs:      factory.InitArgs,
	}
	container.state = &stoppedState{c: &container}
	return &container, nil
}

func (factory *LinuxContainerFactory) Load(id string) (Container, error) {
	panic("implement me")
}

func (factory *LinuxContainerFactory) StartInitialization() error {
	panic("implement me")
}
