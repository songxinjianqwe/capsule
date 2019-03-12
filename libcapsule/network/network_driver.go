package network

type NetworkDriver interface {
	Name() string
	Create(subnet string, name string) (*Network, error)
	Load(name string) (*Network, error)
	Delete(name string) error
	Connect(endpointId string, networkName string, portMappings []string, containerInitPid int) (*Endpoint, error)
	Disconnect(endpoint *Endpoint) error
}
