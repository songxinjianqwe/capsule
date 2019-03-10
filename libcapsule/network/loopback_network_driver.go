package network

import "github.com/songxinjianqwe/capsule/libcapsule/configs"

type LoopbackNetworkDriver struct {
}

func (LoopbackNetworkDriver) Name() string {
	panic("implement me")
}

func (LoopbackNetworkDriver) Create(subnet string, name string) (*Network, error) {
	panic("implement me")
}

func (LoopbackNetworkDriver) Load(name string) (*Network, error) {
	panic("implement me")
}

func (LoopbackNetworkDriver) Delete(*Network) error {
	panic("implement me")
}

func (LoopbackNetworkDriver) Connect(endpointConfig configs.EndpointConfig) (*Endpoint, error) {
	panic("implement me")
}

func (LoopbackNetworkDriver) Disconnect(endpoint *Endpoint) error {
	panic("implement me")
}
