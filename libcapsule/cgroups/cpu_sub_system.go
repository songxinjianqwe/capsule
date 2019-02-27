package cgroups

import "github.com/songxinjianqwe/rune/libcapsule/configc"

type CpuSubSystem struct {
}

func (CpuSubSystem) Name() string {
	return "cpu"
}

func (CpuSubSystem) Remove(*CgroupData) error {
	panic("implement me")
}

func (CpuSubSystem) JoinCgroup(*CgroupData) error {
	panic("implement me")
}

func (CpuSubSystem) SetConfig(path string, cgroup *configc.CgroupConfig) error {
	panic("implement me")
}
