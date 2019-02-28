package cgroups

import (
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"strconv"
)

type CpuSubsystem struct {
	AbstractSubsystem
}

func (subsys *CpuSubsystem) Name() string {
	return "cpu"
}

func (subsys *CpuSubsystem) SetConfig(cgroupName string, cgroupConfig *configc.CgroupConfig) error {
	if cgroupConfig.CpuShares != 0 {
		if err := subsys.WriteConfigEntry(cgroupName, "cpu.shares", []byte(strconv.FormatUint(cgroupConfig.CpuShares, 10))); err != nil {
			return err
		}
	}
	return nil
}
