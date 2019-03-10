package network

import "github.com/songxinjianqwe/capsule/libcapsule/configs"

type BridgeNetworkDriver struct {
}

func (BridgeNetworkDriver) Name() string {
	panic("implement me")
}

func (BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	panic("implement me")
}

func (BridgeNetworkDriver) Load(name string) (*Network, error) {
	panic("implement me")
}

func (BridgeNetworkDriver) Delete(*Network) error {
	panic("implement me")
}

func (BridgeNetworkDriver) Connect(endpointConfig configs.EndpointConfig) (*Endpoint, error) {
	panic("implement me")
}

func (BridgeNetworkDriver) Disconnect(endpoint *Endpoint) error {
	panic("implement me")
}
