package spec

import (
	"github.com/satori/go.uuid"
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
	"github.com/songxinjianqwe/capsule/libcapsule/network"
)

func createNetworkConfig(config *configs.ContainerConfig, networkName string, portMappings []string) error {
	// veth端点
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	if networkName == "" {
		networkName = network.DefaultBridgeName
	}
	config.Endpoint = configs.EndpointConfig{
		ID:           id.String(),
		NetworkName:  networkName,
		PortMappings: portMappings,
	}
	return nil
}
