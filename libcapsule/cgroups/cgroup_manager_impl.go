package cgroups

import (
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"sync"
)

func NewCroupManager(config *configc.Cgroup) CgroupManager {
	return &LinuxCgroupManager{
		Config: config,
		Paths:  make(map[string]string),
	}
}

type LinuxCgroupManager struct {
	mutex  sync.Mutex
	Config *configc.Cgroup
	Paths  map[string]string
}

func (LinuxCgroupManager) Apply(pid int) error {
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
