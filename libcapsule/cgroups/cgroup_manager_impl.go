package cgroups

import (
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"sync"
)

func NewCroupManager() CgroupManager {
	return &LinuxCgroupManager{
		Paths: make(map[string]string),
	}
}

type LinuxCgroupManager struct {
	mutex sync.Mutex
	Paths map[string]string // key是sub system的名称，value是当前容器在该sub system中的路径
}

func (m *LinuxCgroupManager) JoinCgroupSet(pid int) error {
	logrus.Infof("LinuxCgroupManager apply pid: %d", pid)
	for _, subSystem := range subSystems {
		subSystem.JoinCgroup()
	}
	return nil
}

func (m *LinuxCgroupManager) Destroy() error {
	return nil
}

func (m *LinuxCgroupManager) GetPaths() map[string]string {
	return m.Paths
}

func (m *LinuxCgroupManager) SetConfig(config *configc.Config) error {
	return nil
}
