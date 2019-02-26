package spec

import (
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"github.com/songxinjianqwe/rune/libcapsule/util"
)

func createCgroupConfig(cgroupName string, spec *specs.Spec) (*configc.Cgroup, error) {
	logrus.Infof("creating cgroup config...")
	var (
		myCgroupPath string
	)
	c := &configc.Cgroup{
		Resources: &configc.Resources{},
	}

	if spec.Linux != nil && spec.Linux.CgroupsPath != "" {
		myCgroupPath = util.CleanPath(spec.Linux.CgroupsPath)
	}

	if myCgroupPath == "" {
		c.Name = cgroupName
	}
	c.Path = myCgroupPath

	// In rootless containers, any attempt to make cgroup changes is likely to fail.
	// libcapsule will validate this but ignores the error.
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
			dd := &configc.Device{
				Type:        dt,
				Major:       major,
				Minor:       minor,
				Permissions: d.Access,
				Allow:       d.Allow,
			}
			c.Resources.Devices = append(c.Resources.Devices, dd)
		}
		if r.Memory != nil {
			if r.Memory.Limit != nil {
				c.Resources.Memory = *r.Memory.Limit
			}
			if r.Memory.Reservation != nil {
				c.Resources.MemoryReservation = *r.Memory.Reservation
			}
			if r.Memory.Swap != nil {
				c.Resources.MemorySwap = *r.Memory.Swap
			}
			if r.Memory.Kernel != nil {
				c.Resources.KernelMemory = *r.Memory.Kernel
			}
			if r.Memory.KernelTCP != nil {
				c.Resources.KernelMemoryTCP = *r.Memory.KernelTCP
			}
			if r.Memory.Swappiness != nil {
				c.Resources.MemorySwappiness = r.Memory.Swappiness
			}
			if r.Memory.DisableOOMKiller != nil {
				c.Resources.OomKillDisable = *r.Memory.DisableOOMKiller
			}
		}
		if r.CPU != nil {
			if r.CPU.Shares != nil {
				c.Resources.CpuShares = *r.CPU.Shares
			}
			if r.CPU.Quota != nil {
				c.Resources.CpuQuota = *r.CPU.Quota
			}
			if r.CPU.Period != nil {
				c.Resources.CpuPeriod = *r.CPU.Period
			}
			if r.CPU.RealtimeRuntime != nil {
				c.Resources.CpuRtRuntime = *r.CPU.RealtimeRuntime
			}
			if r.CPU.RealtimePeriod != nil {
				c.Resources.CpuRtPeriod = *r.CPU.RealtimePeriod
			}
			if r.CPU.Cpus != "" {
				c.Resources.CpusetCpus = r.CPU.Cpus
			}
			if r.CPU.Mems != "" {
				c.Resources.CpusetMems = r.CPU.Mems
			}
		}
		if r.Pids != nil {
			c.Resources.PidsLimit = r.Pids.Limit
		}
		if r.BlockIO != nil {
			if r.BlockIO.Weight != nil {
				c.Resources.BlkioWeight = *r.BlockIO.Weight
			}
			if r.BlockIO.LeafWeight != nil {
				c.Resources.BlkioLeafWeight = *r.BlockIO.LeafWeight
			}
			if r.BlockIO.WeightDevice != nil {
				for _, wd := range r.BlockIO.WeightDevice {
					var weight, leafWeight uint16
					if wd.Weight != nil {
						weight = *wd.Weight
					}
					if wd.LeafWeight != nil {
						leafWeight = *wd.LeafWeight
					}
					weightDevice := configc.NewWeightDevice(wd.Major, wd.Minor, weight, leafWeight)
					c.Resources.BlkioWeightDevice = append(c.Resources.BlkioWeightDevice, weightDevice)
				}
			}
			if r.BlockIO.ThrottleReadBpsDevice != nil {
				for _, td := range r.BlockIO.ThrottleReadBpsDevice {
					rate := td.Rate
					throttleDevice := configc.NewThrottleDevice(td.Major, td.Minor, rate)
					c.Resources.BlkioThrottleReadBpsDevice = append(c.Resources.BlkioThrottleReadBpsDevice, throttleDevice)
				}
			}
			if r.BlockIO.ThrottleWriteBpsDevice != nil {
				for _, td := range r.BlockIO.ThrottleWriteBpsDevice {
					rate := td.Rate
					throttleDevice := configc.NewThrottleDevice(td.Major, td.Minor, rate)
					c.Resources.BlkioThrottleWriteBpsDevice = append(c.Resources.BlkioThrottleWriteBpsDevice, throttleDevice)
				}
			}
			if r.BlockIO.ThrottleReadIOPSDevice != nil {
				for _, td := range r.BlockIO.ThrottleReadIOPSDevice {
					rate := td.Rate
					throttleDevice := configc.NewThrottleDevice(td.Major, td.Minor, rate)
					c.Resources.BlkioThrottleReadIOPSDevice = append(c.Resources.BlkioThrottleReadIOPSDevice, throttleDevice)
				}
			}
			if r.BlockIO.ThrottleWriteIOPSDevice != nil {
				for _, td := range r.BlockIO.ThrottleWriteIOPSDevice {
					rate := td.Rate
					throttleDevice := configc.NewThrottleDevice(td.Major, td.Minor, rate)
					c.Resources.BlkioThrottleWriteIOPSDevice = append(c.Resources.BlkioThrottleWriteIOPSDevice, throttleDevice)
				}
			}
		}
		for _, l := range r.HugepageLimits {
			c.Resources.HugetlbLimit = append(c.Resources.HugetlbLimit, &configc.HugepageLimit{
				Pagesize: l.Pagesize,
				Limit:    l.Limit,
			})
		}
		if r.Network != nil {
			if r.Network.ClassID != nil {
				c.Resources.NetClsClassid = *r.Network.ClassID
			}
			for _, m := range r.Network.Priorities {
				c.Resources.NetPrioIfpriomap = append(c.Resources.NetPrioIfpriomap, &configc.IfPrioMap{
					Interface: m.Name,
					Priority:  int64(m.Priority),
				})
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

var allowedDevices = []*configc.Device{
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
