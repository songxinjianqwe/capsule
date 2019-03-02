package spec

import (
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
)

var allowedDevices = []*configs.Device{
	// allow mknod for any device
	{
		Type:        'c',
		Major:       wildcard,
		Minor:       wildcard,
		Permissions: "m",
		Allow:       true,
	},
	{
		Type:        'b',
		Major:       wildcard,
		Minor:       wildcard,
		Permissions: "m",
		Allow:       true,
	},
	{
		Type:        'c',
		Path:        "/dev/null",
		Major:       1,
		Minor:       3,
		Permissions: "rwm",
		Allow:       true,
	},
	{
		Type:        'c',
		Path:        "/dev/random",
		Major:       1,
		Minor:       8,
		Permissions: "rwm",
		Allow:       true,
	},
	{
		Type:        'c',
		Path:        "/dev/full",
		Major:       1,
		Minor:       7,
		Permissions: "rwm",
		Allow:       true,
	},
	{
		Type:        'c',
		Path:        "/dev/tty",
		Major:       5,
		Minor:       0,
		Permissions: "rwm",
		Allow:       true,
	},
	{
		Type:        'c',
		Path:        "/dev/zero",
		Major:       1,
		Minor:       5,
		Permissions: "rwm",
		Allow:       true,
	},
	{
		Type:        'c',
		Path:        "/dev/urandom",
		Major:       1,
		Minor:       9,
		Permissions: "rwm",
		Allow:       true,
	},
	{
		Path:        "/dev/console",
		Type:        'c',
		Major:       5,
		Minor:       1,
		Permissions: "rwm",
		Allow:       true,
	},
	// /dev/pts/ - pts namespaces are "coming soon"
	{
		Path:        "",
		Type:        'c',
		Major:       136,
		Minor:       wildcard,
		Permissions: "rwm",
		Allow:       true,
	},
	{
		Path:        "",
		Type:        'c',
		Major:       5,
		Minor:       2,
		Permissions: "rwm",
		Allow:       true,
	},
	// tuntap
	{
		Path:        "",
		Type:        'c',
		Major:       10,
		Minor:       200,
		Permissions: "rwm",
		Allow:       true,
	},
}

func createCgroupConfig(spec *specs.Spec) (*configs.Cgroup, error) {
	logrus.Infof("creating cgroup config...")
	c := &configs.Cgroup{
		Resources: &configs.Resources{},
	}

	if spec.Linux != nil {
		r := spec.Linux.Resources
		if r == nil {
			return c, nil
		}
		for i, d := range spec.Linux.Resources.Devices {
			var (
				t     = "a"
				major = int64(-1)
				minor = int64(-1)
			)
			if d.Type != "" {
				t = d.Type
			}
			if d.Major != nil {
				major = *d.Major
			}
			if d.Minor != nil {
				minor = *d.Minor
			}
			if d.Access == "" {
				return nil, fmt.Errorf("device access at %d field cannot be empty", i)
			}
			dt, err := stringToCgroupDeviceRune(t)
			if err != nil {
				return nil, err
			}
			device := &configs.Device{
				Type:        dt,
				Major:       major,
				Minor:       minor,
				Permissions: d.Access,
				Allow:       d.Allow,
			}
			c.Resources.Devices = append(c.Resources.Devices, device)
		}
		if r.Memory != nil {
			if r.Memory.Limit != nil {
				c.Resources.Memory = *r.Memory.Limit
			}
		}
		if r.CPU != nil {
			if r.CPU.Shares != nil {
				c.Resources.CpuShares = *r.CPU.Shares
			}
			if r.CPU.Cpus != "" {
				c.Resources.CpusetCpus = r.CPU.Cpus
			}
		}
	}
	// append the default allowed devices to the end of the list
	c.Resources.Devices = append(c.Resources.Devices, allowedDevices...)
	return c, nil
}

func stringToCgroupDeviceRune(s string) (rune, error) {
	switch s {
	case "a":
		return 'a', nil
	case "b":
		return 'b', nil
	case "c":
		return 'c', nil
	default:
		return 0, fmt.Errorf("invalid cgroup device type %q", s)
	}
}
