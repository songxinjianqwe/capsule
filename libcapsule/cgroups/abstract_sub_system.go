package cgroups

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type AbstractSubsystem struct {
}

func (subsys *AbstractSubsystem) Name() string {
	panic("implement me")
}

func (subsys *AbstractSubsystem) Remove(cgroupName string) error {
	cgroupPath, err := createAndGetCgroupAbsolutePathIfNotExists(subsys.Name(), cgroupName)
	if err != nil {
		return err
	}
	if err := os.Remove(cgroupPath); err != nil {
		return err
	}
	return nil
}

func (subsys *AbstractSubsystem) Join(cgroupName string, pid int) (string, error) {
	cgroupPath, err := createAndGetCgroupAbsolutePathIfNotExists(subsys.Name(), cgroupName)
	if err != nil {
		return "", err
	}
	// write pid
	if err := ioutil.WriteFile(
		path.Join(cgroupPath, "tasks"),
		[]byte(strconv.Itoa(pid)),
		0644); err != nil {
		return "", err
	}
	return cgroupPath, nil
}

func (subsys *AbstractSubsystem) WriteConfigEntry(cgroupName, configFilename string, data []byte) error {
	cgroupPath, err := createAndGetCgroupAbsolutePathIfNotExists(subsys.Name(), cgroupName)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(path.Join(cgroupPath, configFilename),
		data, 0644); err != nil {
		return err
	}
	return nil
}
