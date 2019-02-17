package libcapsule

import (
	"bytes"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/songxinjianqwe/rune/libcapsule/cgroups"
	"github.com/songxinjianqwe/rune/libcapsule/config"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const ()

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
	panic("implement me")
}

func (c *LinuxContainer) State() (*State, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.currentState()
}

func (c *LinuxContainer) currentState() (*State, error) {
	panic("implement me")
}

func (c *LinuxContainer) OCIState() (*specs.State, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.currentOCIState()
}

func (c *LinuxContainer) currentOCIState() (*specs.State, error) {
	panic("implement me")
}

func (c *LinuxContainer) runStatus() (Status, error) {
	if c.initProcess == nil {
		return Stopped, nil
	}
	pid := c.initProcess.pid()
	stat, err := util.Stat(pid)
	if err != nil {
		return Stopped, nil
	}
	if stat.StartTime != c.initProcessStartTime || stat.State == util.Zombie || stat.State == util.Dead {
		return Stopped, nil
	}
	// We'll create exec fifo and blocking on it after container is created,
	// and delete it after start container.
	if _, err := os.Stat(filepath.Join(c.root, execFifoFilename)); err == nil {
		return Created, nil
	}
	return Running, nil
}

func (c *LinuxContainer) Config() config.Config {
	panic("implement me")
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

func (c *LinuxContainer) isPaused() (bool, error) {
	fcg := c.cgroupManager.GetPaths()["freezer"]
	if fcg == "" {
		// A container doesn't have a freezer cgroup
		return false, nil
	}
	data, err := ioutil.ReadFile(filepath.Join(fcg, "freezer.state"))
	if err != nil {
		// If freezer cgroup is not mounted, the container would just be not paused.
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, util.NewGenericError(err, util.SystemError)
	}
	return bytes.Equal(bytes.TrimSpace(data), []byte("FROZEN")), nil
}

// refreshState needs to be called to verify that the current state on the
// container is what is true.  Because consumers of libcontainer can use it
// out of process we need to verify the container's status based on runtime
// information and not rely on our in process info.
func (c *LinuxContainer) refreshState() error {
	paused, err := c.isPaused()
	if err != nil {
		return err
	}
	if paused {
		return c.state.transition(&pausedState{c: c})
	}
	t, err := c.runStatus()
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
