package libcapsule

import (
	"github.com/opencontainers/runc/libcontainer/cgroups"
	"github.com/songxinjianqwe/rune/libcapsule/config"
	"github.com/songxinjianqwe/rune/libcapsule/config/validate"
)

type LinuxContainerFactory struct {
	// Root directory for the factory to store state.
	Root string

	// InitPath is the path for calling the init responsibilities for spawning
	// a container.
	InitPath string

	// InitArgs are arguments for calling the init responsibilities for spawning
	// a container.
	InitArgs []string

	// New{u,g}uidmapPath is the path to the binaries used for mapping with
	// rootless containers.
	NewuidmapPath string
	NewgidmapPath string

	// Validator provides validation to container configurations.
	Validator validate.Validator

	// NewCgroupsManager returns an initialized cgroups manager for a single container.
	NewCgroupsManager func(config *config.Cgroup, paths map[string]string) cgroups.Manager
}

func (factory *LinuxContainerFactory) Create(id string, config *config.Config) (Container, error) {
	panic("implement me")
}

func (factory *LinuxContainerFactory) Load(id string) (Container, error) {
	panic("implement me")
}

func (factory *LinuxContainerFactory) StartInitialization() error {
	panic("implement me")
}
