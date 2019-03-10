package network

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
)

/*
对应一个网段，Driver取值有Bridge
*/
type Network struct {
	// 网络名称
	Name string `json:"name"`
	// 网段
	IpRange *net.IPNet `json:"ip_range"`
	// 网络驱动名（网络类型）
	Driver string `json:"driver"`
}

/*
对应一个网络端点，比如容器中会有一个veth和一个loopback
*/
type Endpoint struct {
	ID           string           `json:"id"`
	IpAddress    net.IP           `json:"ip_address"`
	MacAddress   net.HardwareAddr `json:"mac_address"`
	Device       netlink.Veth     `json:"-"`
	Network      *Network         `json:"network"`
	PortMappings []string         `json:"port_mappings"`
}

/*
如果receiver是指针类型，则接口值必须为指针；如果receiver均为值类型，则接口值可以是指针，也可以是值。
一点规则：有值，未必能取得指针；反之一定可以。
*/
var networkDrivers = map[string]NetworkDriver{
	"bridge":   &BridgeNetworkDriver{},
	"loopback": &LoopbackNetworkDriver{},
}

func CreateNetwork(driver string, subnet string, name string) (*Network, error) {
	networkDriver, found := networkDrivers[driver]
	if !found {
		return nil, fmt.Errorf("network driver not found: %s", driver)
	}
	// subnet的格式是192.168.1.0/24，parseCIDR的第一个返回值是IP地址，第二个返回值是IPNet类型，192.168.1.0/24
	_, ipRange, err := net.ParseCIDR(subnet)
	if err != nil {
		return nil, err
	}
	allocator, err := LoadIPAllocator()
	if err != nil {
		return nil, err
	}
	gatewayIP, err := allocator.Allocate(ipRange)
	if err != nil {
		return nil, err
	}
	ipRange.IP = gatewayIP
	// 即使设置了IP，网段也不会发生变化，因为网段地址是 IP&subnet mask(取决于/24等)
	return networkDriver.Create(ipRange.String(), name)
}

func DeleteNetwork(driver string, name string) error {
	networkDriver, found := networkDrivers[driver]
	if !found {
		return fmt.Errorf("network driver not found: %s", driver)
	}
	network, err := LoadNetwork(driver, name)
	if err != nil {
		return err
	}
	allocator, err := LoadIPAllocator()
	if err != nil {
		return err
	}
	if err := allocator.Release(network.IpRange, &network.IpRange.IP); err != nil {
		return err
	}
	return networkDriver.Delete(network)
}

func LoadNetwork(driver string, name string) (*Network, error) {
	networkDriver, found := networkDrivers[driver]
	if !found {
		return nil, fmt.Errorf("network driver not found: %s", driver)
	}
	return networkDriver.Load(name)
}

func Connect(networkDriver string, endpointId string, networkName string, portMappings []string) (*Endpoint, error) {
	networkDriverInstance, found := networkDrivers[networkDriver]
	if !found {
		return nil, fmt.Errorf("network driver not found: %s", networkDriver)
	}
	return networkDriverInstance.Connect(endpointId, networkName, portMappings)
}

func Disconnect(endpoint *Endpoint) error {
	networkDriver, found := networkDrivers[endpoint.Network.Driver]
	if !found {
		return fmt.Errorf("network driver not found: %s", endpoint.Network.Driver)
	}
	return networkDriver.Disconnect(endpoint)
}
