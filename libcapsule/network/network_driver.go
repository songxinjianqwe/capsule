package network

import "github.com/songxinjianqwe/capsule/libcapsule/configs"

type NetworkDriver interface {
	Name() string
	Create(subnet string, name string) (*configs.Network, error)
	Delete(*configs.Network) error
	Connect(network *configs.Network, endpoint *configs.Endpoint) error
	Disconnect(network *configs.Network, endpoint *configs.Endpoint) error
}
