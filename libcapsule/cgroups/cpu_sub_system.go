package cgroups

import (
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"strconv"
)

type CpuSubsystem struct {
}

func (subsys *CpuSubsystem) Name() string {
	return "cpu"
}

func (subsys *CpuSubsystem) SetConfig(cgroupName string, cgroupConfig *configc.CgroupConfig) error {
	logrus.Infof("process is setting config in [%s] subsystem", subsys.Name())
	if cgroupConfig.CpuShares != 0 {
		logrus.Infof("writing config, cpushares: %d", cgroupConfig.CpuShares)
		if err := writeConfigEntry(subsys.Name(), cgroupName, "cpu.shares", []byte(strconv.FormatUint(cgroupConfig.CpuShares, 10))); err != nil {
			return err
		}
	}
	return nil
}
