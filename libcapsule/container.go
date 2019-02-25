package libcapsule

import (
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"os"
)

// Container is a libcapsule container object.
// Each container is thread-safe within the same config. Since a container can
// be destroyed by a separate config, any function may return that the container
// was not found.
type Container interface {
	ID() string

	// errors:
	// ContainerNotExists - Container no longer exists,
	// Systemerror - System util.
	Status() (ContainerStatus, error)

	// errors:
	// SystemError - System util.
	State() (*StateStorage, error)

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

	// 创建但不运行cmd
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

	// 删除容器，如果容器还处于created，则杀掉容器进程
	// errors:
	// ContainerNotStopped - Container is still running,
	// ContainerPaused - Container is paused,
	// SystemError - System util.
	Destroy() error

	// 向容器init进程发送信号
	// errors:
	// SystemError - System util.
	Signal(s os.Signal) error

	// 让容器执行最终命令
	// errors:
	// SystemError - System util.
	Start() error
}
