package cgroups

import (
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
)

type CpusetSubsystem struct {
}

func (subsys *CpusetSubsystem) Name() string {
	return "cpuset"
}

func (subsys *CpusetSubsystem) SetConfig(cgroupName string, cgroupConfig *configc.CgroupConfig) error {
	if cgroupConfig.CpusetCpus != "" {
		logrus.Infof("config is cpuset cpus: %d", cgroupConfig.CpusetCpus)
		if err := writeConfigEntry(subsys.Name(), cgroupName, "cpuset.cpus", []byte(cgroupConfig.CpusetCpus)); err != nil {
			return err
		}
	}
	return nil
}
