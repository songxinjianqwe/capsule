package cgroups

import (
	"github.com/songxinjianqwe/rune/libcapsule/configc"
)

type CgroupManager interface {
	// Applies cgroup configuration to the process with the specified pid
	Apply(pid int) error

	// Destroys the cgroup set
	Destroy() error

	// The option func SystemdCgroups() and Cgroupfs() require following attributes:
	// 	Paths   map[string]string
	// 	CgroupConfig *configs.CgroupConfig
	// Paths maps cgroup subsystem to path at which it is mounted.
	// CgroupConfig specifies specific cgroup settings for the various subsystems

	// Returns cgroup paths to save in a state file and to be able to
	// restore the object later.
	GetPaths() map[string]string

	// Sets the cgroup as configured.
	Set(container *configc.Config) error
}
