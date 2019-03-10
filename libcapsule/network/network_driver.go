package network

import "github.com/songxinjianqwe/capsule/libcapsule/configs"

type NetworkDriver interface {
	Name() string
	Create(subnet string, name string) (*Network, error)
	Load(name string) (*Network, error)
	Delete(*Network) error
	Connect(endpointConfig configs.EndpointConfig) (*Endpoint, error)
	Disconnect(endpoint *Endpoint) error
}
