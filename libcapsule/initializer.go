package libcapsule

import (
	"fmt"
	"os"
)

type InitializerType string

const (
	SetnsInitializer    InitializerType = "setns"
	StandardInitializer InitializerType = "standard"
)

type Initializer interface {
	Init() error
}

func NewInitializer(initializerType InitializerType, config *InitConfig, childPipe *os.File, execFifoFd int) (Initializer, error) {
	switch initializerType {
	case StandardInitializer:
		return &InitializerStandardImpl{
			config:     config,
			childPipe:  childPipe,
			execFifoFd: execFifoFd,
		}, nil
	case SetnsInitializer:
		return &InitializerSetnsImpl{
			config:    config,
			childPipe: childPipe,
		}, nil
	default:
		return nil, fmt.Errorf("unknown initializerType:%s", initializerType)
	}
}
