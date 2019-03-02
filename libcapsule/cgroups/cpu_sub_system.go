package cgroups

import (
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
	"strconv"
)

type CpuSubsystem struct {
}

func (subsys *CpuSubsystem) Name() string {
	return "cpu"
}

/**
cpu share，默认值为1024。
多个容器都设置了cpu share，并且每个容器的进程都会把CPU沾满时：
将每个容器的cpu share的值相加，每个容器的占比就是 CPU 的利用率。
如果只有一个容器，那么此时它无论设置 512 或者 1024，CPU 利用率都将是 100%。
*/
func (subsys *CpuSubsystem) SetConfig(cgroupName string, cgroupConfig *configs.Cgroup) error {
	logrus.Infof("process is setting config in [%s] subsystem", subsys.Name())
	if cgroupConfig.CpuShares != 0 {
		logrus.Infof("writing config, cpushares: %d", cgroupConfig.CpuShares)
		if err := writeConfigEntry(subsys.Name(), cgroupName, "cpu.shares", []byte(strconv.FormatUint(cgroupConfig.CpuShares, 10))); err != nil {
			return err
		}
	}
	return nil
}
