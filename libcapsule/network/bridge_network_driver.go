package network

import "github.com/songxinjianqwe/capsule/libcapsule/configs"

type BridgeNetworkDriver struct {
}

func (BridgeNetworkDriver) Name() string {
	panic("implement me")
}

func (BridgeNetworkDriver) Create(subnet string, name string) (*configs.Network, error) {
	panic("implement me")
}

func (BridgeNetworkDriver) Delete(*configs.Network) error {
	panic("implement me")
}

func (BridgeNetworkDriver) Connect(network *configs.Network, endpoint *configs.Endpoint) error {
	panic("implement me")
}

func (BridgeNetworkDriver) Disconnect(network *configs.Network, endpoint *configs.Endpoint) error {
	panic("implement me")
}
