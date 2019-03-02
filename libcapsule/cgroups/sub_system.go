package cgroups

import (
	"github.com/songxinjianqwe/capsule/libcapsule/configc"
)

/**
以memory为例：
memory是一个subsystem。
/sys/fs/cgroup/memory是一个hierarchy（或者称为mount point ），它与memory subsystem关联。
这个目录是hierarchy的root节点。
我们要做的是在这个目录下创建一个目录（子节点），称为一个cgroup，并在该目录下的tasks文件中写入pid，将该进程加入到这个cgroup中。
Join方法传入的为cgroup的name
getCgroupAbsolutePath就是将subsystem name 映射为 相应的hierarchy root，然后与cgroup name拼接起来。
*/
type Subsystem interface {
	SubsystemSpecific
	SubsystemCommon
}

type SubsystemSpecific interface {
	// Name returns the name of the subsystem.
	Name() string
	// Set the cgroup represented by cgroup.
	SetConfig(cgroupName string, cgroupConfig *configc.Cgroup) error
}

type SubsystemCommon interface {
	// Removes the cgroup
	Remove(cgroupName string) error
	// Creates and joins the cgroup
	Join(cgroupName string, pid int) (string, error)
}

var (
	subSystems = []Subsystem{
		&SubsystemWrapper{
			child: &CpuSubsystem{},
		},
		&SubsystemWrapper{
			child: &MemorySubsystem{},
		},
	}
)
