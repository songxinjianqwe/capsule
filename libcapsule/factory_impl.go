package libcapsule

import (
	"github.com/songxinjianqwe/rune/libcapsule/cgroups"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"github.com/songxinjianqwe/rune/libcapsule/configc/validate"
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
	factory := LinuxContainerFactory{
		Root:      RuntimeRoot,
		Validator: validate.New(),
		// TODO
		NewCgroupsManager: nil,
	}
	return &factory, nil
}

type LinuxContainerFactory struct {
	// Root directory for the factory to store containerState.
	Root string

	// Validator provides validation to container configurations.
	Validator validate.Validator

	// NewCgroupsManager returns an initialized cgroups manager for a single container.
	NewCgroupsManager func(config *configc.Cgroup, paths map[string]string) cgroups.CgroupManager
}

func (factory *LinuxContainerFactory) Create(id string, config *configc.Config) (Container, error) {
	container := &LinuxContainer{
		id:            id,
		root:          factory.Root,
		config:        *config,
		cgroupManager: factory.NewCgroupsManager(config.Cgroups, nil),
	}
	container.containerState = &StoppedState{c: container}
	return container, nil
}

func (factory *LinuxContainerFactory) Load(id string) (Container, error) {
	panic("implement me")
}

func (factory *LinuxContainerFactory) StartInitialization() error {
	panic("implement me")
}
