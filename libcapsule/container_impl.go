package libcapsule

import (
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/songxinjianqwe/rune/libcapsule/cgroups"
	"github.com/songxinjianqwe/rune/libcapsule/config"
	"github.com/songxinjianqwe/rune/libcapsule/util/proc"
	"golang.org/x/sys/unix"
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
	initProcess          ParentProcess
	initProcessStartTime uint64
	state                containerState
	created              time.Time
	mutex                sync.Mutex
}

// ************************************************************************************************
// public
// ************************************************************************************************

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

func (c *LinuxContainer) Start(process *Process) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.start(process)
}

func (c *LinuxContainer) Run(process *Process) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if err := c.start(process); err != nil {
		return err
	}
	if err := c.exec(); err != nil {
		return err
	}
	return nil
}

func (c *LinuxContainer) Destroy() error {
	panic("implement me")
}

func (c *LinuxContainer) Signal(s os.Signal, all bool) error {
	panic("implement me")
}

func (c *LinuxContainer) Exec() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.exec()
}

// ************************************************************************************************
// private
// ************************************************************************************************
func (c *LinuxContainer) start(process *Process) (err error) {
	if err := c.createExecFifo(); err != nil {
		return err
	}
	panic("implement me")
	c.deleteExecFifo()
	return nil
}

func (c *LinuxContainer) exec() error {
	panic("implement me")
}

func (c *LinuxContainer) currentState() (*State, error) {
	panic("implement me")
}

func (c *LinuxContainer) currentOCIState() (*specs.State, error) {
	panic("implement me")
}

/**
1、如果容器init进程不存在，或者进程已经死亡或成为僵尸进程，则均为 【Stopped】
2、如果exec.fifo文件存在，则为 【Created】
3、其他情况为 【Running】
*/
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
	// 在容器创建前，会先创建execfifo管道；在容器创建后，会删除该管道
	if _, err := os.Stat(filepath.Join(c.root, execFifoFilename)); err == nil {
		return Created, nil
	}
	return Running, nil
}

/**
在start前，创建exec.fifo管道
io.Pipe是内存管道，无法通过内存管道来感知容器状态
因为管道存在，则说明容器是处于created之后，running之前的状态
*/
func (c *LinuxContainer) createExecFifo() error {
	fifoName := filepath.Join(c.root, execFifoFilename)

	if _, err := os.Stat(fifoName); err == nil {
		return fmt.Errorf("exec fifo %s already exists", fifoName)
	}
	// 读是4，写是2，执行是1
	// 自己可以读写，同组可以写，其他组可以写
	if err := unix.Mkfifo(fifoName, 0622); err != nil {
		return err
	}
	return nil
}

/**
在start后，删除exec.fifo管道
*/
func (c *LinuxContainer) deleteExecFifo() error {
	fifoName := filepath.Join(c.root, execFifoFilename)
	return os.Remove(fifoName)
}

/**
创建一个parent process，用于与容器init进程通信
*/
func (c *LinuxContainer) newParentProcess() (ParentProcess, error) {
	panic("implement me")
}
