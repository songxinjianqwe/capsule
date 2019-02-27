package cgroups

import (
	"github.com/songxinjianqwe/rune/libcapsule/configc"
)

type MemorySubSystem struct {
}

func (MemorySubSystem) Name() string {
	panic("implement me")
}

func (MemorySubSystem) Remove(*CgroupData) error {
	panic("implement me")
}

func (MemorySubSystem) Apply(*CgroupData) error {
	panic("implement me")
}

func (MemorySubSystem) Set(path string, cgroup *configc.CgroupConfig) error {
	panic("implement me")
}
