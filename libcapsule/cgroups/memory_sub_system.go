package cgroups

import (
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"strconv"
)

type MemorySubsystem struct {
}

func (subsys *MemorySubsystem) Name() string {
	return "memory"
}

func (subsys *MemorySubsystem) SetConfig(cgroupName string, cgroupConfig *configc.CgroupConfig) error {
	if cgroupConfig.Memory > 0 {
		logrus.Infof("config is memory: %d", cgroupConfig.Memory)
		if err := writeConfigEntry(subsys.Name(), cgroupName, "memory.limit_in_bytes", []byte(strconv.FormatInt(cgroupConfig.Memory, 10))); err != nil {
			return err
		}
	}
	return nil
}
