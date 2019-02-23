package libcapsule

import (
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"os"
	"time"
)

// Status is the status of a container.
type Status int

const (
	// Created is the status that denotes the container exists but has not been run yet.
	Created Status = iota
	// Running is the status that denotes the container exists and is running.
	Running
	// Stopped is the status that denotes the container does not have a created or running process.
	Stopped
)

func (s Status) String() string {
	switch s {
	case Created:
		return "Created"
	case Running:
		return "Running"
	case Stopped:
		return "Stopped"
	default:
		return "Unknown"
	}
}

func statusFromString(s string) Status {
	switch s {
	case "Created":
		return Created
	case "Running":
		return Running
	case "Stopped":
		return Stopped
	default:
		return -1
	}
}

// State represents a running container's state
type State struct {
	// ID is the container ID.
	ID string `json:"id"`

	// Status is the container Status
	Status string `json:"status"`

	// InitProcessPid is the init process id in the parent namespace.
	InitProcessPid int `json:"init_process_pid"`

	// InitProcessStartTime is the init process start time in clock cycles since boot time.
	InitProcessStartTime uint64 `json:"init_process_start"`

	// Created is the unix timestamp for the creation time of the container in UTC
	Created time.Time `json:"create_time"`

	// Config is the container's configuration.
	Config configc.Config `json:"config"`

	// Path to all the cgroups setup for a container. Key is cgroup subsystem name
	// with the value as the path.
	CgroupPaths map[string]string `json:"cgroup_paths"`

	// NamespacePaths are filepaths to the container's namespaces. Key is the namespace type
	// with the value as the path.
	NamespacePaths map[configc.NamespaceType]string `json:"namespace_paths"`
}

// Container is a libcapsule container object.
// Each container is thread-safe within the same process. Since a container can
// be destroyed by a separate process, any function may return that the container
// was not found.
type Container interface {
	ID() string

	// errors:
	// ContainerNotExists - Container no longer exists,
	// Systemerror - System util.
	Status() (Status, error)

	// errors:
	// SystemError - System util.
	State() (*State, error)

	// errors:
	// SystemError - System util.
	OCIState() (*specs.State, error)

	// Returns the current configc of the container.
	Config() configc.Config

	// 返回容器内的PIDs，存放在namespace中
	// errors:
	// ContainerNotExists - Container no longer exists,
	// Systemerror - System util.
	//
	// Some of the returned PIDs may no longer refer to processes in the Container, unless
	// the Container state is PAUSED in which case every PID in the slice is valid.
	Processes() ([]int, error)

	// 创建但不运行
	// errors:
	// ContainerNotExists - Container no longer exists,
	// ConfigInvalid - configc is invalid,
	// ContainerPaused - Container is paused,
	// SystemError - System util.
	Create(process *Process) (err error)

	// Create + Start
	// errors:
	// ContainerNotExists - Container no longer exists,
	// ConfigInvalid - configc is invalid,
	// ContainerPaused - Container is paused,
	// SystemError - System util.
	Run(process *Process) (err error)

	// 在杀掉所有运行进程后，销毁容器
	// errors:
	// ContainerNotStopped - Container is still running,
	// ContainerPaused - Container is paused,
	// SystemError - System util.
	Destroy() error

	// 向容器init进程发送信号
	// errors:
	// SystemError - System util.
	Signal(s os.Signal, all bool) error

	// 让容器执行最终命令
	// errors:
	// SystemError - System util.
	Start() error
}
