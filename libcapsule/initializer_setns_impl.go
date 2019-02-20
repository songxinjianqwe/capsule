package libcapsule

import "os"

type InitializerSetnsImpl struct {
	config    *InitConfig
	childPipe *os.File
}

func (initializer *InitializerSetnsImpl) Init() error {
	panic("implement me")
}
