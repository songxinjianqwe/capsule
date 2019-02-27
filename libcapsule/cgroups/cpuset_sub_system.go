package cgroups

import "github.com/songxinjianqwe/rune/libcapsule/configc"

type CpusetSubSystem struct {
}

func (CpusetSubSystem) Name() string {
	return "cpuset"
}

func (CpusetSubSystem) Remove(*CgroupData) error {
	panic("implement me")
}

func (CpusetSubSystem) JoinCgroup(*CgroupData) error {
	panic("implement me")
}

func (CpusetSubSystem) SetConfig(path string, cgroup *configc.CgroupConfig) error {
	panic("implement me")
}
