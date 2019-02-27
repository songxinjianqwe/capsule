package cgroups

import (
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"sync"
)

func NewCroupManager(config *configc.CgroupConfig) CgroupManager {
	return &LinuxCgroupManager{
		Config: config,
		Paths:  make(map[string]string),
	}
}

type LinuxCgroupManager struct {
	mutex  sync.Mutex
	Config *configc.CgroupConfig
	Paths  map[string]string
}

func (m *LinuxCgroupManager) Apply(pid int) error {
	logrus.Infof("LinuxCgroupManager apply pid:%d", pid)
	return nil
}

func (LinuxCgroupManager) GetPids() ([]int, error) {
	return nil, nil
}

func (LinuxCgroupManager) GetAllPids() ([]int, error) {
	return nil, nil
}

func (LinuxCgroupManager) GetStats() (*Stats, error) {
	return nil, nil
}

func (LinuxCgroupManager) Freeze(state configc.FreezerState) error {
	return nil
}

func (LinuxCgroupManager) Destroy() error {
	return nil
}

func (LinuxCgroupManager) GetPaths() map[string]string {
	return nil
}

func (LinuxCgroupManager) Set(container *configc.Config) error {
	return nil
}
