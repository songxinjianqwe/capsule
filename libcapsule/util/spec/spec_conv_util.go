package spec

import (
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"os"
	"path/filepath"
)

/*
将specs.Spec转为libcapsule.Config
*/
func CreateContainerConfig(id string, spec *specs.Spec) (*configc.Config, error) {
	logrus.Infof("converting specs.Spec to libcapsule.Config...")
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
	config := &configc.Config{
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

	// 转换cgroups
	cgroupConfig, err := createCgroupConfig(id, spec)
	if err != nil {
		return nil, err
	}
	config.Cgroups = cgroupConfig
	logrus.Infof("convert cgroup config complete, config.Cgroups: %#v", config.Cgroups)

	// Linux特有配置
	if spec.Linux != nil {
		// 转换namespaces
		for _, ns := range spec.Linux.Namespaces {
			t, exists := namespaceMapping[ns.Type]
			if !exists {
				return nil, fmt.Errorf("namespace %q does not exist", ns)
			}
			if config.Namespaces.Contains(t) {
				return nil, fmt.Errorf("malformed spec file: duplicated ns %q", ns)
			}
			config.Namespaces.Add(t, ns.Path)
		}
		if config.Namespaces.Contains(configc.NEWNET) && config.Namespaces.PathOf(configc.NEWNET) == "" {
			config.Networks = []*configc.Network{
				{
					Type: "loopback",
				},
			}
		}
		logrus.Infof("convert namespaces complete, config.Namespaces: %#v", config.Namespaces)
		config.Sysctl = spec.Linux.Sysctl
	}
	config.Version = specs.Version
	return config, nil
}

const wildcard = -1

var namespaceMapping = map[specs.LinuxNamespaceType]configc.NamespaceType{
	specs.PIDNamespace:     configc.NEWPID,
	specs.NetworkNamespace: configc.NEWNET,
	specs.MountNamespace:   configc.NEWNS,
	specs.UserNamespace:    configc.NEWUSER,
	specs.IPCNamespace:     configc.NEWIPC,
	specs.UTSNamespace:     configc.NEWUTS,
	specs.CgroupNamespace:  configc.NEWCGROUP,
}
