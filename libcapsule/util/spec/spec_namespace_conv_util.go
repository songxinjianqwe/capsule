package spec

import (
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/songxinjianqwe/capsule/libcapsule/configc"
)

const wildcard = -1

var namespaceMapping = map[specs.LinuxNamespaceType]configc.NamespaceType{
	specs.PIDNamespace:     configc.NEWPID,
	specs.NetworkNamespace: configc.NEWNET,
	specs.MountNamespace:   configc.NEWNS,
	specs.IPCNamespace:     configc.NEWIPC,
	specs.UTSNamespace:     configc.NEWUTS,
}

func createNamespaces(config *configc.ContainerConfig, spec *specs.Spec) error {
	// 转换namespaces
	for _, ns := range spec.Linux.Namespaces {
		t, exists := namespaceMapping[ns.Type]
		if !exists {
			return fmt.Errorf("namespace %q does not exist", ns)
		}
		if config.Namespaces.Contains(t) {
			return fmt.Errorf("malformed spec file: duplicated ns %q", ns)
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
	return nil
}
