package cgroups

import (
	"github.com/songxinjianqwe/rune/libcapsule/configc"
)

type SubSystem interface {
	// Name returns the name of the subsystem.
	Name() string
	// Removes the cgroup represented by 'CgroupData'.
	Remove(*CgroupData) error
	// Creates and joins the cgroup represented by 'CgroupData'.
	JoinCgroup(*CgroupData) error
	// Set the cgroup represented by cgroup.
	SetConfig(path string, cgroup *configc.CgroupConfig) error
}

var (
	subsystems = []SubSystem{
		&CpuSubSystem{},
		&MemorySubSystem{},
		&CpusetSubSystem{},
	}
)
