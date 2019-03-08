package spec

import (
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
)

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
	return c, nil
}
