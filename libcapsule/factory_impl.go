package libcapsule

import "github.com/songxinjianqwe/rune/libcapsule/config"

type LinuxContainerFactory struct {
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
