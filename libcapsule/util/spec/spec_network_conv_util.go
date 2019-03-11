package spec

import (
	"github.com/satori/go.uuid"
	"github.com/songxinjianqwe/capsule/libcapsule"
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
)

func createNetworkConfig(config *configs.ContainerConfig, portMappings []string) error {
	// veth端点
	id, err := uuid.NewV4()
	if err != nil {
		return err
	}
	config.Endpoint = configs.EndpointConfig{
		ID:            id.String(),
		NetworkDriver: "bridge",
		NetworkName:   libcapsule.DefaultBridgeName,
		PortMappings:  portMappings,
	}
	return nil
}
