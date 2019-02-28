package cgroups

import (
	"github.com/songxinjianqwe/rune/libcapsule/configc"
)

type CpusetSubsystem struct {
	AbstractSubsystem
}

func (subsys *CpusetSubsystem) Name() string {
	return "cpuset"
}

func (subsys *CpusetSubsystem) SetConfig(cgroupName string, cgroupConfig *configc.CgroupConfig) error {
	if cgroupConfig.CpusetCpus != "" {
		if err := subsys.WriteConfigEntry(cgroupName, "cpuset.cpus", []byte(cgroupConfig.CpusetCpus)); err != nil {
			return err
		}
	}
	return nil
}
