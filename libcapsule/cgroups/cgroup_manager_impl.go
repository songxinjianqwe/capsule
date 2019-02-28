package cgroups

import (
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"sync"
)

func NewCroupManager(id string, paths map[string]string) CgroupManager {
	if paths == nil {
		paths = make(map[string]string)
	}
	return &LinuxCgroupManager{
		CgroupName: id,
		Paths:      paths,
	}
}

type LinuxCgroupManager struct {
	mutex      sync.Mutex
	CgroupName string
	Paths      map[string]string // key是sub system的名称，value是当前容器在该hierarchy中的路径
}

func (m *LinuxCgroupManager) JoinCgroupSet(pid int) (err error) {
	logrus.Infof("process %d is joining cgroup set %s", pid, m.CgroupName)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, subSystem := range subSystems {
		var cgroupPath string
		if cgroupPath, err = subSystem.Join(m.CgroupName, pid); err != nil {
			return err
		}
		m.Paths[subSystem.Name()] = cgroupPath
	}
	return nil
}

func (m *LinuxCgroupManager) Destroy() (err error) {
	logrus.Infof("destroying cgroup set %s", m.CgroupName)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, subSystem := range subSystems {
		if err = subSystem.Remove(m.CgroupName); err != nil {
			logrus.Warnf("remove subsys %s failed, cause: %s", m.CgroupName, err.Error())
		}
	}
	m.Paths = make(map[string]string)
	return err
}

func (m *LinuxCgroupManager) GetPaths() map[string]string {
	return m.Paths
}

func (m *LinuxCgroupManager) SetConfig(cgroupConfig *configc.CgroupConfig) error {
	logrus.Infof("set cgroup set %s config", m.CgroupName)
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for _, subSystem := range subSystems {
		if err := subSystem.SetConfig(m.CgroupName, cgroupConfig); err != nil {
			return err
		}
	}
	return nil
}
