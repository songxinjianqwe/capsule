package configc

type CgroupConfig struct {
	// Paths represent the absolute cgroups paths to join.
	// This takes precedence over Path.
	Paths map[string]string

	// Resources contains various cgroups settings to apply
	// 继承
	*Resources
}

type Resources struct {
	Devices []*Device `json:"devices"`

	// Memory limit (in bytes)
	Memory int64 `json:"memory"`

	// CPU shares (relative weight vs. other containers)
	CpuShares uint64 `json:"cpu_shares"`

	// CPU to use
	CpusetCpus string `json:"cpuset_cpus"`
}
