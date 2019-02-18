package cgroups

import (
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"sync"
)

func NewCroupManager(config *configc.Cgroup, paths map[string]string) CgroupManager {
	return &LinuxCgroupManager{
		Config: config,
		Paths:  paths,
	}
}

type LinuxCgroupManager struct {
	mutex  sync.Mutex
	Config *configc.Cgroup
	Paths  map[string]string
}

func (LinuxCgroupManager) Apply(pid int) error {
	panic("implement me")
}

func (LinuxCgroupManager) GetPids() ([]int, error) {
	panic("implement me")
}

func (LinuxCgroupManager) GetAllPids() ([]int, error) {
	panic("implement me")
}

func (LinuxCgroupManager) GetStats() (*Stats, error) {
	panic("implement me")
}

func (LinuxCgroupManager) Freeze(state configc.FreezerState) error {
	panic("implement me")
}

func (LinuxCgroupManager) Destroy() error {
	panic("implement me")
}

func (LinuxCgroupManager) GetPaths() map[string]string {
	panic("implement me")
}

func (LinuxCgroupManager) Set(container *configc.Config) error {
	panic("implement me")
}
