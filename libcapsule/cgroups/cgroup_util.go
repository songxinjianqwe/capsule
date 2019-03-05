package cgroups

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"strings"
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
			logrus.Infof("cgroup path not found, then create it: %s", cgroupAbsolutePath)
			if err := os.Mkdir(cgroupAbsolutePath, 0644); err != nil {
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
	// example:
	// 29 26 0:26 / /sys/fs/cgroup/memory rw,nosuid,nodev,noexec,relatime shared:10 - cgroup cgroup rw,seclabel,memory
	//
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

func writeConfigEntry(subsystemName, cgroupName, configFilename string, data []byte) error {
	cgroupPath, err := createAndGetCgroupAbsolutePathIfNotExists(subsystemName, cgroupName)
	if err != nil {
		return err
	}
	logrus.Infof("write to [%s]: %s", path.Join(cgroupPath, configFilename), string(data))
	if err := ioutil.WriteFile(path.Join(cgroupPath, configFilename),
		data, 0644); err != nil {
		return err
	}
	return nil
}
