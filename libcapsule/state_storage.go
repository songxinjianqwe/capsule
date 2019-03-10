package libcapsule

import (
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
	"github.com/songxinjianqwe/capsule/libcapsule/network"
	"time"
)

// StateStorage represents a running container's state
type StateStorage struct {
	// ID is the container ID.
	ID string `json:"id"`

	// InitProcessPid is the init process id in the parent namespace.
	InitProcessPid int `json:"init_process_pid"`

	// InitProcessStartTime is the init process create time in clock cycles since boot time.
	InitProcessStartTime uint64 `json:"init_process_start_time"`

	// Created is the unix timestamp for the creation time of the container in UTC
	Created time.Time `json:"create_time"`

	// ContainerConfig is the container's configuration.
	Config configs.ContainerConfig `json:"config"`

	// Path to all the cgroups setup for a container. Key is cgroup subsystem name
	// with the value as the path.
	CgroupPaths map[string]string `json:"cgroup_paths"`

	// NamespacePaths are filepaths to the container's namespaces. Key is the namespace type
	// with the value as the path.
	NamespacePaths map[configs.NamespaceType]string `json:"namespace_paths"`

	// Endpoints
	Endpoints []network.Endpoint `json:"endpoints"`
}
