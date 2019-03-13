package spec

import (
	"github.com/satori/go.uuid"
	"github.com/songxinjianqwe/capsule/libcapsule"
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
)

func createNetworkConfig(config *configs.ContainerConfig, network string, portMappings []string) error {
	// veth端点
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	if network == "" {
		network = libcapsule.DefaultBridgeName
	}
	config.Endpoint = configs.EndpointConfig{
		ID:           id.String(),
		NetworkName:  network,
		PortMappings: portMappings,
	}
	return nil
}
