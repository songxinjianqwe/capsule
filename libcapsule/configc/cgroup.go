package configc

type FreezerState string

const (
	Undefined FreezerState = ""
	Frozen    FreezerState = "FROZEN"
	Thawed    FreezerState = "THAWED"
)

type CgroupConfig struct {
	// Deprecated, use Path instead
	Name string `json:"name,omitempty"`

	// name of parent of cgroup or slice
	// Deprecated, use Path instead
	Parent string `json:"parent,omitempty"`

	// Path specifies the path to cgroups that are created and/or joined by the container.
	// The path is assumed to be relative to the host system cgroup mountpoint.
	Path string `json:"path"`

	// ScopePrefix describes prefix for the scope name
	ScopePrefix string `json:"scope_prefix"`

	// Paths represent the absolute cgroups paths to join.
	// This takes precedence over Path.
	Paths map[string]string

	// Resources contains various cgroups settings to apply
	*Resources
}

type Resources struct {
	Devices []*Device `json:"devices"`

	// Memory limit (in bytes)
	Memory int64 `json:"memory"`

	// CPU shares (relative weight vs. other containers)
	CpuShares uint64 `json:"cpu_shares"`

	// set the freeze value for the process
	Freezer FreezerState `json:"freezer"`
}
