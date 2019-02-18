package libcapsule

import (
	"github.com/songxinjianqwe/rune/libcapsule/config"
)

type Factory interface {
	// 创建一个新的容器
	// errors:
	// IdInUse - id is already in use by a container
	// InvalidIdFormat - id has incorrect format
	// ConfigInvalid - config is invalid
	// Systemerror - System util
	//
	// On util, any partially createdTime container parts are cleaned up (the operation is atomic).
	Create(id string, config *config.Config) (Container, error)

	// 加载一个容器
	// errors:
	// Path does not exist
	// System util
	Load(id string) (Container, error)

	// 用于init进程初始化
	// Errors:
	// Pipe connection util
	// System util
	StartInitialization() error
}
