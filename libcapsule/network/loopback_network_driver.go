package network

import "github.com/songxinjianqwe/capsule/libcapsule/configs"

type LoopbackNetworkDriver struct {
}

func (LoopbackNetworkDriver) Name() string {
	panic("implement me")
}

func (LoopbackNetworkDriver) Create(subnet string, name string) (*configs.Network, error) {
	panic("implement me")
}

func (LoopbackNetworkDriver) Delete(*configs.Network) error {
	panic("implement me")
}

func (LoopbackNetworkDriver) Connect(network *configs.Network, endpoint *configs.Endpoint) error {
	panic("implement me")
}

func (LoopbackNetworkDriver) Disconnect(network *configs.Network, endpoint *configs.Endpoint) error {
	panic("implement me")
}
