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
	Apply(*CgroupData) error
	// Set the cgroup represented by cgroup.
	Set(path string, cgroup *configc.CgroupConfig) error
}

type SubSystemSet []SubSystem

var (
	subsystems = SubSystemSet{
		&CpuSubSystem{},
		&MemorySubSystem{},
		&CpusetSubSystem{},
	}
)
