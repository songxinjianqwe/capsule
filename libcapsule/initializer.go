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

func NewInitializer(initializerType InitializerType, config *InitConfig, configPipe *os.File) (Initializer, error) {
	switch initializerType {
	case StandardInitializer:
		return &InitializerStandardImpl{
			config:     config,
			configPipe: configPipe,
		}, nil
	case SetnsInitializer:
		return &InitializerSetnsImpl{
			config:    config,
			childPipe: configPipe,
		}, nil
	default:
		return nil, fmt.Errorf("unknown initializerType:%s", initializerType)
	}
}
