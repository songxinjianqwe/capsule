package libcapsule

import (
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/songxinjianqwe/rune/libcapsule/cgroups"
	"github.com/songxinjianqwe/rune/libcapsule/config"
	"github.com/songxinjianqwe/rune/libcapsule/util/proc"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type LinuxContainer struct {
	id                   string
	root                 string
	config               config.Config
	cgroupManager        cgroups.CgroupManager
	initPath             string
	initArgs             []string
	initProcess          parentProcess
	initProcessStartTime uint64
	state                containerState
	created              time.Time
	mutex                sync.Mutex
}

func (c *LinuxContainer) ID() string {
	return c.id
}

func (c *LinuxContainer) Status() (Status, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.currentStatus()
}

func (c *LinuxContainer) State() (*State, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.currentState()
}

func (c *LinuxContainer) OCIState() (*specs.State, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.currentOCIState()
}

func (c *LinuxContainer) Config() config.Config {
	return c.config
}

func (c *LinuxContainer) Processes() ([]int, error) {
	panic("implement me")
}

func (c *LinuxContainer) Start(process *Process) (err error) {
	panic("implement me")
}

func (c *LinuxContainer) Run(process *Process) (err error) {
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

// ************************************************************************************************
// private
// ************************************************************************************************

func (c *LinuxContainer) currentState() (*State, error) {
	panic("implement me")
}

func (c *LinuxContainer) currentOCIState() (*specs.State, error) {
	panic("implement me")
}

func (c *LinuxContainer) currentStatus() (Status, error) {
	if c.initProcess == nil {
		return Stopped, nil
	}
	pid := c.initProcess.pid()
	stat, err := proc.Stat(pid)
	if err != nil {
		return Stopped, nil
	}
	if stat.StartTime != c.initProcessStartTime || stat.State == proc.Zombie || stat.State == proc.Dead {
		return Stopped, nil
	}
	// We'll create exec fifo and blocking on it after container is created,
	// and delete it after start container.
	if _, err := os.Stat(filepath.Join(c.root, execFifoFilename)); err == nil {
		return Created, nil
	}
	return Running, nil
}

// refreshState needs to be called to verify that the current state on the
// container is what is true.  Because consumers of libcontainer can use it
// out of process we need to verify the container's status based on runtime
// information and not rely on our in process info.
func (c *LinuxContainer) refreshState() error {
	t, err := c.currentStatus()
	if err != nil {
		return err
	}
	switch t {
	case Created:
		return c.state.transition(&createdState{c: c})
	case Running:
		return c.state.transition(&runningState{c: c})
	}
	return c.state.transition(&stoppedState{c: c})
}
