package libcapsule

import (
	"os"
	"runtime"
)

func NewInitializer(config *InitConfig, childPipe *os.File) Initializer {
	return &InitializerImpl{
		config:    config,
		childPipe: childPipe,
	}
}

type InitializerImpl struct {
	config    *InitConfig
	childPipe *os.File
}

func (initializer *InitializerImpl) Init() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := initializer.setUpNetwork(); err != nil {
		return err
	}
	if err := initializer.setUpRoute(); err != nil {
		return err
	}

}

func (initializer *InitializerImpl) setUpNetwork() error {

}

func (initializer *InitializerImpl) setUpRoute() error {

}
