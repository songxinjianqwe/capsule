package network

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
	"strings"
)

type BridgeNetworkDriver struct {
}

func (driver *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (driver *BridgeNetworkDriver) Create(subnet string, name string) (*Network, error) {
	gatewayIP, ipRange, err := net.ParseCIDR(subnet)
	network := &Network{
		Name:    name,
		IpRange: ipRange,
		Driver:  driver.Name(),
	}
	// subnet的格式是192.168.1.2/24，parseCIDR的第一个返回值是IP地址,192.168.1.2，第二个返回值是IPNet类型，192.168.1.0/24
	if err != nil {
		return nil, err
	}
	// 1.创建bridge
	if err := createBridgeInterface(name); err != nil {
		return nil, err
	}
	// 2.设置Bridge的IP地址和路由
	if err := setInterfaceIP(name, gatewayIP.String(), ipRange); err != nil {

	}
	return network, nil
}

func setInterfaceIP(name string, gatewayIP string, subnet *net.IPNet) error {
	return nil
}

func (driver *BridgeNetworkDriver) Load(name string) (*Network, error) {
	panic("implement me")
}

func (driver *BridgeNetworkDriver) Delete(*Network) error {
	panic("implement me")
}

func (driver *BridgeNetworkDriver) Connect(endpointId string, networkName string, portMappings []string) (*Endpoint, error) {
	network, err := LoadNetwork(driver.Name(), networkName)
	if err != nil {
		return nil, err
	}
	allocator, err := LoadIPAllocator()
	if err != nil {
		return nil, err
	}
	ip, err := allocator.Allocate(network.IpRange)
	if err != nil {
		return nil, err
	}
	endpoint := &Endpoint{
		ID:           endpointId,
		Network:      network,
		IpAddress:    ip,
		PortMappings: portMappings,
	}
	// connect
	// config ip address and route
	// config port mapping
	return endpoint, nil
}

func (driver *BridgeNetworkDriver) Disconnect(endpoint *Endpoint) error {
	panic("implement me")
}

// ******************************************************************************************
// util
// ******************************************************************************************

func createBridgeInterface(name string) error {
	if _, err := net.InterfaceByName(name); err == nil || strings.Contains(err.Error(), "no such network interface") {
		return fmt.Errorf("brigde name %s exists", name)
	}
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = name
	br := &netlink.Bridge{
		LinkAttrs: linkAttrs,
	}
	if err := netlink.LinkAdd(br); err != nil {
		return fmt.Errorf("bridge create failed, cause: %s", err.Error())
	}
	return nil
}
