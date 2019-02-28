package cgroups

import (
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"strconv"
)

type MemorySubsystem struct {
	AbstractSubsystem
}

func (subsys *MemorySubsystem) Name() string {
	return "memory"
}

func (subsys *MemorySubsystem) SetConfig(cgroupName string, cgroupConfig *configc.CgroupConfig) error {
	if cgroupConfig.Memory > 0 {
		if err := subsys.WriteConfigEntry(cgroupName, "memory.limit_in_bytes", []byte(strconv.FormatInt(cgroupConfig.Memory, 10))); err != nil {
			return err
		}
	}
	return nil
}
