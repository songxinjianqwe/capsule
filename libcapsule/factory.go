package libcapsule

import (
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
)

type Factory interface {
	// 创建一个新的容器
	Create(id string, config *configs.ContainerConfig) (Container, error)

	// 加载一个容器
	Load(id string) (Container, error)

	Exists(id string) bool

	// 用于init/exec进程初始化
	StartInitialization() error

	// 返回运行时文件的根目录
	GetRuntimeRoot() string
}
