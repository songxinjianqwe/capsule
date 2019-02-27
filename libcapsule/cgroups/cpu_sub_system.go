package cgroups

import "github.com/songxinjianqwe/rune/libcapsule/configc"

type CpuSubSystem struct {
}

func (CpuSubSystem) Name() string {
	panic("implement me")
}

func (CpuSubSystem) Remove(*CgroupData) error {
	panic("implement me")
}

func (CpuSubSystem) Apply(*CgroupData) error {
	panic("implement me")
}

func (CpuSubSystem) Set(path string, cgroup *configc.CgroupConfig) error {
	panic("implement me")
}
