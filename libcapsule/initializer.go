package libcapsule

import (
	"fmt"
	"github.com/songxinjianqwe/capsule/libcapsule/constant"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
)

type InitializerType string

const (
	ExecInitializer InitializerType = "exec"
	InitInitializer InitializerType = "init"
)

type Initializer interface {
	Init() error
}

func NewInitializer(initializerType InitializerType, config *InitExecConfig, configPipe *os.File, runtimeRoot string) (Initializer, error) {
	containerRoot := filepath.Join(runtimeRoot, constant.ContainerDir, config.ID)
	switch initializerType {
	case InitInitializer:
		return &InitializerStandardImpl{
			config:        config,
			configPipe:    configPipe,
			parentPid:     unix.Getppid(),
			containerRoot: containerRoot,
		}, nil
	case ExecInitializer:
		return &InitializerExecImpl{
			config:        config,
			configPipe:    configPipe,
			containerRoot: containerRoot,
		}, nil
	default:
		return nil, fmt.Errorf("unknown initializerType:%s", initializerType)
	}
}
