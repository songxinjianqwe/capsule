package libcapsule

import (
	"fmt"
	"github.com/songxinjianqwe/rune/libcapsule/cgroups"
	"os"
	"os/exec"
)

func NewParentSetnsProcess(process *Process, cmd *exec.Cmd, parentConfigPipe *os.File) ParentProcess {
	cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", EnvInitializerType, string(SetnsInitializer)))
	return &ParentSetnsProcess{
		initProcessCmd:   cmd,
		parentConfigPipe: parentConfigPipe,
		process:          process,
	}
}

/**
ProcessWrapper接口的实现类，包裹了SetnsProcess
*/
type ParentSetnsProcess struct {
	initProcessCmd   *exec.Cmd
	parentConfigPipe *os.File
	process          *Process
	cgroupManger     *cgroups.CgroupManager
}

func (p *ParentSetnsProcess) detach() bool {
	return p.process.Detach
}

func (p *ParentSetnsProcess) pid() int {
	panic("implement me")
}

func (p *ParentSetnsProcess) start() error {
	panic("implement me")
}

func (p *ParentSetnsProcess) terminate() error {
	panic("implement me")
}

func (p *ParentSetnsProcess) wait() error {
	panic("implement me")
}

func (p *ParentSetnsProcess) startTime() (uint64, error) {
	panic("implement me")
}

func (p *ParentSetnsProcess) signal(os.Signal) error {
	panic("implement me")
}
