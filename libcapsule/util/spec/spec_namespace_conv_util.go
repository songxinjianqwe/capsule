package spec

import (
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
)

var namespaceMapping = map[specs.LinuxNamespaceType]configs.NamespaceType{
	specs.PIDNamespace:     configs.NEWPID,
	specs.NetworkNamespace: configs.NEWNET,
	specs.MountNamespace:   configs.NEWNS,
	specs.IPCNamespace:     configs.NEWIPC,
	specs.UTSNamespace:     configs.NEWUTS,
}

func createNamespaces(config *configs.ContainerConfig, spec *specs.Spec) error {
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
	if config.Namespaces.Contains(configs.NEWNET) && config.Namespaces.PathOf(configs.NEWNET) == "" {
		config.Endpoints = []*configs.Endpoint{
			{
				Type: "loopback",
			},
		}
	}
	return nil
}
