package cgroups

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"os"
	"path"
	"strings"
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
	// Name returns the name of the subsystem.
	Name() string
	// Removes the cgroup
	Remove(cgroupName string) error
	// Creates and joins the cgroup
	Join(cgroupName string, pid int) (string, error)
	// Set the cgroup represented by cgroup.
	SetConfig(cgroupName string, cgroupConfig *configc.CgroupConfig) error
}

var (
	subSystems = []Subsystem{
		&CpuSubsystem{},
		&MemorySubsystem{},
		&CpusetSubsystem{},
	}
)

func createAndGetCgroupAbsolutePathIfNotExists(subsystemName string, cgroupName string) (string, error) {
	hierarchyRoot, err := findCgroupMountpoint(subsystemName)
	if err != nil {
		return "", err
	}
	cgroupAbsolutePath := path.Join(hierarchyRoot, cgroupName)
	if _, err := os.Stat(cgroupAbsolutePath); err != nil {
		// 目录不存在，则创建
		if os.IsNotExist(err) {
			if err := os.Mkdir(cgroupAbsolutePath, 0755); err != nil {
				logrus.Errorf("create cgroup relative path %s failed, cause: %s", cgroupAbsolutePath, err.Error())
				return "", err
			}
		} else {
			// 目录坏掉了，返回失败
			return "", err
		}
	}
	// 原本就存在
	return cgroupAbsolutePath, nil
}

func findCgroupMountpoint(subsystemName string) (string, error) {
	// cat /proc/self/mountinfo 拿到当前进程的相关mount信息
	// 29 26 0:26 / /sys/fs/cgroup/memory rw,nosuid,nodev,noexec,relatime shared:10 - cgroup cgroup rw,seclabel,memory
	// 文件里都是这种格式一行一行存储的
	// 注意，最后是rw,memory，而memory是我们的subsystemName
	// 按照空格split的话，我们需要的路径信息为[4]
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return "", nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, " ")
		lastField := fields[len(fields)-1]
		lastFieldFields := strings.Split(lastField, ",")
		for _, value := range lastFieldFields {
			if value == subsystemName {
				return fields[4], nil
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("subsys %s's cgroup mountpoint not found", subsystemName)
}
