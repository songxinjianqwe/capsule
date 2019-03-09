package spec

import (
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
	"os"
	"path/filepath"
)

/*
将specs.Spec转为libcapsule.ContainerConfig
*/
func CreateContainerConfig(spec *specs.Spec) (*configs.ContainerConfig, error) {
	logrus.Infof("converting specs.Spec to libcapsule.ContainerConfig...")
	// runc's cwd will always be the bundle path
	rcwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	// 拿到当前路径，即bundle path
	cwd, err := filepath.Abs(rcwd)
	if err != nil {
		return nil, err
	}
	if spec.Root == nil {
		return nil, fmt.Errorf("root must be specified")
	}

	// rootfs path要么是绝对路径，要么是cwd+rootfs转为绝对路径
	rootfsPath := spec.Root.Path
	if !filepath.IsAbs(rootfsPath) {
		rootfsPath = filepath.Join(cwd, rootfsPath)
	}
	logrus.Infof("rootfs path is %s", rootfsPath)

	// 将annotations转为labels
	var labels []string
	for k, v := range spec.Annotations {
		labels = append(labels, fmt.Sprintf("%s=%s", k, v))
	}
	config := &configs.ContainerConfig{
		Rootfs:     rootfsPath,
		Readonlyfs: spec.Root.Readonly,
		Hostname:   spec.Hostname,
		Labels:     append(labels, fmt.Sprintf("bundle=%s", cwd)),
	}

	// 转换挂载
	for _, m := range spec.Mounts {
		mount := createMount(cwd, m)
		logrus.Infof("convert mount complete: %#v", mount)
		config.Mounts = append(config.Mounts, mount)
	}
	logrus.Infof("convert mounts complete, config.Mounts: %#v", config.Mounts)

	// 转换设备
	if err := createDevices(spec, config); err != nil {
		return nil, err
	}
	logrus.Infof("convert devices complete, config.Devices: %#v", config.Devices)

	// 转换cgroup
	cgroupConfig, err := createCgroupConfig(spec)
	if err != nil {
		return nil, err
	}
	config.Cgroup = cgroupConfig
	logrus.Infof("convert cgroup config complete, config.Cgroup: %#v", config.Cgroup)

	// 转换网络
	for _, m := range spec.Mounts {
		mount := createMount(cwd, m)
		logrus.Infof("convert mount complete: %#v", mount)
		config.Mounts = append(config.Mounts, mount)
	}

	// Linux特有配置
	if spec.Linux != nil {
		if err := createNamespaces(config, spec); err != nil {
			return nil, err
		}
		logrus.Infof("convert namespaces complete, config.Namespaces: %#v", config.Namespaces)
		config.Sysctl = spec.Linux.Sysctl
	}
	config.Version = specs.Version
	return config, nil
}
