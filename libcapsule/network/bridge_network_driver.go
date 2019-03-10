package network

type BridgeNetworkDriver struct {
}

func (driver *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (driver *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	panic("implement me")
}

func (driver *BridgeNetworkDriver) Load(name string) (*Network, error) {
	panic("implement me")
}

func (driver *BridgeNetworkDriver) Delete(*Network) error {
	panic("implement me")
}

func (driver *BridgeNetworkDriver) Connect(endpointId string, networkName string, portMappings []string) (*Endpoint, error) {
	panic("implement me")
}

func (driver *BridgeNetworkDriver) Disconnect(endpoint *Endpoint) error {
	panic("implement me")
}
