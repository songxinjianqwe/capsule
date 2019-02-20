package libcapsule

import (
	"fmt"
	"github.com/songxinjianqwe/rune/libcapsule/cgroups"
	"os"
	"os/exec"
)

func NewSetnsProcessWrapper(process *Process, cmd *exec.Cmd, parentConfigPipe *os.File) ProcessWrapper {
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", EnvInitializerType, string(SetnsInitializer)))
	return &SetnsProcessWrapperImpl{
		initProcessCmd:   cmd,
		parentConfigPipe: parentConfigPipe,
		process:          process,
	}
}

/**
ProcessWrapper接口的实现类，包裹了SetnsProcess
*/
type SetnsProcessWrapperImpl struct {
	initProcessCmd   *exec.Cmd
	parentConfigPipe *os.File
	process          *Process
	cgroupManger     *cgroups.CgroupManager
}

func (p *SetnsProcessWrapperImpl) pid() int {
	panic("implement me")
}

func (p *SetnsProcessWrapperImpl) start() error {
	panic("implement me")
}

func (p *SetnsProcessWrapperImpl) terminate() error {
	panic("implement me")
}

func (p *SetnsProcessWrapperImpl) wait() (*os.ProcessState, error) {
	panic("implement me")
}

func (p *SetnsProcessWrapperImpl) startTime() (uint64, error) {
	panic("implement me")
}

func (p *SetnsProcessWrapperImpl) signal(os.Signal) error {
	panic("implement me")
}
