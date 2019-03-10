package network

type BridgeNetworkDriver struct {
}

func (driver *BridgeNetworkDriver) Name() string {
	panic("implement me")
}

func (driver *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	panic("implement me")
}

func (driver *BridgeNetworkDriver) Delete(*Network) error {
	panic("implement me")
}

func (driver *BridgeNetworkDriver) Connect(network *Network, endpoint *Endpoint) error {
	panic("implement me")
}

func (driver *BridgeNetworkDriver) Disconnect(network *Network, endpoint *Endpoint) error {
	panic("implement me")
}
