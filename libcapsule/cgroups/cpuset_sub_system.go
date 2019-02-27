package cgroups

import "github.com/songxinjianqwe/rune/libcapsule/configc"

type CpusetSubSystem struct {
}

func (CpusetSubSystem) Name() string {
	panic("implement me")
}

func (CpusetSubSystem) Remove(*CgroupData) error {
	panic("implement me")
}

func (CpusetSubSystem) Apply(*CgroupData) error {
	panic("implement me")
}

func (CpusetSubSystem) Set(path string, cgroup *configc.CgroupConfig) error {
	panic("implement me")
}
