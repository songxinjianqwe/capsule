package libcapsule

import (
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/songxinjianqwe/rune/libcapsule/config"
	"github.com/songxinjianqwe/rune/libcapsule/process"
	"os"
)

type LinuxContainer struct {
}

func (c *LinuxContainer) ID() string {
	panic("implement me")
}

func (c *LinuxContainer) Status() (Status, error) {
	panic("implement me")
}

func (c *LinuxContainer) State() (*State, error) {
	panic("implement me")
}

func (c *LinuxContainer) OCIState() (*specs.State, error) {
	panic("implement me")
}

func (c *LinuxContainer) Config() config.Config {
	panic("implement me")
}

func (c *LinuxContainer) Processes() ([]int, error) {
	panic("implement me")
}

func (c *LinuxContainer) Set(config config.Config) error {
	panic("implement me")
}

func (c *LinuxContainer) Start(process *process.Process) (err error) {
	panic("implement me")
}

func (c *LinuxContainer) Run(process *process.Process) (err error) {
	panic("implement me")
}

func (c *LinuxContainer) Destroy() error {
	panic("implement me")
}

func (c *LinuxContainer) Signal(s os.Signal, all bool) error {
	panic("implement me")
}

func (c *LinuxContainer) Exec() error {
	panic("implement me")
}
