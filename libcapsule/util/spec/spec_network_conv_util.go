package spec

import (
	"github.com/satori/go.uuid"
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
)

func createNetworkConfig(config *configs.ContainerConfig, portMappings []string) error {
	if config.Namespaces.Contains(configs.NEWNET) && config.Namespaces.PathOf(configs.NEWNET) == "" {
		id, err := uuid.NewV4()
		if err != nil {
			return err
		}
		config.Endpoints = append(config.Endpoints, &configs.EndpointConfig{
			ID:   id.String(),
			Type: "loopback",
		})
	}
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	config.Endpoints = append(config.Endpoints, &configs.EndpointConfig{
		ID:           id.String(),
		Type:         "veth",
		PortMappings: portMappings,
	})
	return nil
}
