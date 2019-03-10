package network

type NetworkDriver interface {
	Name() string
	Create(subnet string, name string) (*Network, error)
	Delete(*Network) error
	Connect(network *Network, endpoint *Endpoint) error
	Disconnect(network *Network, endpoint *Endpoint) error
}
