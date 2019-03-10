package network

type LoopbackNetworkDriver struct {
}

func (driver *LoopbackNetworkDriver) Name() string {
	return "loopback"
}

func (driver *LoopbackNetworkDriver) Create(subnet string, name string) (*Network, error) {
	panic("implement me")
}

func (driver *LoopbackNetworkDriver) Load(name string) (*Network, error) {
	panic("implement me")
}

func (driver *LoopbackNetworkDriver) Delete(*Network) error {
	panic("implement me")
}

func (driver *LoopbackNetworkDriver) Connect(endpointId string, networkName string, portMappings []string) (*Endpoint, error) {
	panic("implement me")
}

func (driver *LoopbackNetworkDriver) Disconnect(endpoint *Endpoint) error {
	panic("implement me")
}
