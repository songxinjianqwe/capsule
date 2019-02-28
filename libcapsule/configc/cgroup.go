package configc

type CgroupConfig struct {
	// Path specifies the path to cgroups that are created and/or joined by the container.
	// The path is assumed to be relative to the host system cgroup mountpoint.
	// root cgroup目录的当前cgroup节点的相对路径
	Path string `json:"path"`

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
