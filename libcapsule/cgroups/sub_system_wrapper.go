package cgroups

import (
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

/**
模板方法模式
父类组合子类
如果使用子类继承父类的方式，那么在调用父类的Name方法时，会跳到父类的Name去执行，无法跳到子类重写的的Name。
*/
type SubsystemWrapper struct {
	child SubsystemSpecific
}

func (subsys *SubsystemWrapper) Name() string {
	return subsys.child.Name()
}

func (subsys *SubsystemWrapper) SetConfig(cgroupName string, cgroupConfig *configc.CgroupConfig) error {
	return subsys.child.SetConfig(cgroupName, cgroupConfig)
}

func (subsys *SubsystemWrapper) Remove(cgroupName string) error {
	cgroupPath, err := createAndGetCgroupAbsolutePathIfNotExists(subsys.Name(), cgroupName)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(cgroupPath); err != nil {
		return err
	}
	return nil
}

func (subsys *SubsystemWrapper) Join(cgroupName string, pid int) (string, error) {
	logrus.Infof("process is joining %s subsystem", subsys.Name())
	cgroupPath, err := createAndGetCgroupAbsolutePathIfNotExists(subsys.Name(), cgroupName)
	if err != nil {
		return "", err
	}
	// write pid
	logrus.Infof("writing pid [%d] to %s", pid, path.Join(cgroupPath, "tasks"))
	if err := ioutil.WriteFile(
		path.Join(cgroupPath, "tasks"),
		[]byte(strconv.Itoa(pid)),
		0700); err != nil {
		return "", err
	}
	return cgroupPath, nil
}
