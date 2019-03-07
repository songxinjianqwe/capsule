package libcapsule

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
)

type InitializerType string

const (
	ExecInitializer     InitializerType = "exec"
	StandardInitializer InitializerType = "standard"
)

type Initializer interface {
	Init() error
}

func NewInitializer(initializerType InitializerType, config *InitConfig, configPipe *os.File) (Initializer, error) {
	switch initializerType {
	case StandardInitializer:
		return &InitializerStandardImpl{
			config:     config,
			configPipe: configPipe,
			parentPid:  unix.Getppid(),
		}, nil
	case ExecInitializer:
		return nil, fmt.Errorf("exec initializer cant be used for now")
		//return &InitializerExecImpl{
		//	config:    config,
		//	childPipe: configPipe,
		//}, nil
	default:
		return nil, fmt.Errorf("unknown initializerType:%s", initializerType)
	}
}
