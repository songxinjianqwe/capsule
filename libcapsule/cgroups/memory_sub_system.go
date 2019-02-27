package cgroups

import (
	"github.com/songxinjianqwe/rune/libcapsule/configc"
)

type MemorySubSystem struct {
}

func (MemorySubSystem) Name() string {
	return "memory"
}

func (MemorySubSystem) Remove(*CgroupData) error {
	panic("implement me")
}

func (MemorySubSystem) JoinCgroup(*CgroupData) error {
	panic("implement me")
}

func (MemorySubSystem) SetConfig(path string, cgroup *configc.CgroupConfig) error {
	panic("implement me")
}
