package spec

import (
	"github.com/satori/go.uuid"
	"github.com/songxinjianqwe/capsule/libcapsule"
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
)

func createNetworkConfig(config *configs.ContainerConfig, portMappings []string) error {
	if config.Namespaces.Contains(configs.NEWNET) && config.Namespaces.PathOf(configs.NEWNET) == "" {
		id, err := uuid.NewV4()
		if err != nil {
			return err
		}
		config.Endpoints = append(config.Endpoints, configs.EndpointConfig{
			ID:            id.String(),
			NetworkDriver: "loopback",
		})
	}
	// veth端点
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	config.Endpoints = append(config.Endpoints, configs.EndpointConfig{
		ID:            id.String(),
		NetworkDriver: "bridge",
		NetworkName:   libcapsule.DefaultBridgeName,
		PortMappings:  portMappings,
	})
	return nil
}
