package libcapsule

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
)

type InitializerType string

const (
	ExecInitializer InitializerType = "exec"
	InitInitializer InitializerType = "init"
)

type Initializer interface {
	Init() error
}

func NewInitializer(initializerType InitializerType, config *InitExecConfig, configPipe *os.File) (Initializer, error) {
	switch initializerType {
	case InitInitializer:
		return &InitializerStandardImpl{
			config:     config,
			configPipe: configPipe,
			parentPid:  unix.Getppid(),
		}, nil
	case ExecInitializer:
		return &InitializerExecImpl{
			config:     config,
			configPipe: configPipe,
		}, nil
	default:
		return nil, fmt.Errorf("unknown initializerType:%s", initializerType)
	}
}
