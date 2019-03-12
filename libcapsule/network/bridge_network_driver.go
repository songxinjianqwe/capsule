package network

import (
	"fmt"
	"github.com/coreos/go-iptables/iptables"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
	"github.com/vishvananda/netlink"
	"net"
)

const Masquerade = "-t nat -A POSTROUTING -s %s -o %s -j MASQUERADE"

type BridgeNetworkDriver struct {
}

func (driver *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (driver *BridgeNetworkDriver) Create(subnet string, bridgeName string) (*Network, error) {
	_, ipRange, err := net.ParseCIDR(subnet)
	network := &Network{
		Name:    bridgeName,
		IpRange: *ipRange,
		Driver:  driver.Name(),
	}
	// subnet的格式是192.168.1.2/24，parseCIDR的第一个返回值是IP地址,192.168.1.2，第二个返回值是IPNet类型，192.168.1.0/24
	if err != nil {
		return nil, err
	}
	// 1.创建bridge
	if err := createBridgeInterface(bridgeName); err != nil {
		return nil, exception.NewGenericErrorWithContext(err, exception.SystemError, "create bridge")
	}

	// 2.设置Bridge的IP地址和路由
	if err := setInterfaceIPAndRoute(bridgeName, subnet); err != nil {
		return nil, exception.NewGenericErrorWithContext(err, exception.SystemError, "set bridge ip and route")
	}

	// 3.启动Bridge
	if err := setInterfaceUp(bridgeName); err != nil {
		return nil, exception.NewGenericErrorWithContext(err, exception.SystemError, "set bridge UP")
	}

	// 4.设置iptables SNAT规则（MASQUERADE）
	if err := setupIPTablesMasquerade(bridgeName, *ipRange); err != nil {
		return nil, exception.NewGenericErrorWithContext(err, exception.SystemError, "set iptables SNAT MASQUERADE RULE")
	}
	return network, nil
}

func (driver *BridgeNetworkDriver) Load(name string) (*Network, error) {
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return nil, err
	}
	//  `ip addr show`.
	addrs, err := netlink.AddrList(iface, netlink.FAMILY_ALL)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("addresses not found")
	}
	var bridgeAddr *net.IPNet
	for _, addr := range addrs {
		if addr.Label == name {
			bridgeAddr = addr.IPNet
			break
		}
	}
	if bridgeAddr == nil {
		return nil, fmt.Errorf("addresses not found")
	}
	return &Network{
		Name:    name,
		Driver:  driver.Name(),
		IpRange: *bridgeAddr,
	}, nil
}

func (driver *BridgeNetworkDriver) Delete(name string) error {
	network, err := driver.Load(name)
	if err != nil {
		return err
	}
	// 删除SNAT规则
	tables, err := iptables.New()
	if err := tables.Delete(
		"nat",
		"POSTROUTING",
		getSNATRuleSpecs(network.Name, network.IpRange)...,
	); err != nil {
		return err
	}
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}
	if err := netlink.LinkDel(iface); err != nil {
		return err
	}
	return nil
}

func (driver *BridgeNetworkDriver) Connect(endpointId string, networkName string, portMappings []string, containerInitPid int) (*Endpoint, error) {
	network, err := LoadNetwork(driver.Name(), networkName)
	if err != nil {
		return nil, exception.NewGenericErrorWithContext(err, exception.SystemError, "load network")
	}
	allocator, err := LoadIPAllocator()
	if err != nil {
		return nil, exception.NewGenericErrorWithContext(err, exception.SystemError, "load ip allocator")
	}
	endpointIP, err := allocator.Allocate(&network.IpRange)
	if err != nil {
		return nil, exception.NewGenericErrorWithContext(err, exception.SystemError, "allocate ip")
	}
	endpoint := &Endpoint{
		Name:         endpointId,
		Network:      network,
		IpAddress:    endpointIP,
		PortMappings: portMappings,
	}
	logrus.Infof("connecting network, endpoint: %#v, veth ip: %s", endpoint, endpoint.IpAddress.String())
	// 创建网络端点veth
	if err := createVethPairAndSetUp(endpoint); err != nil {
		return nil, exception.NewGenericErrorWithContext(err, exception.SystemError, "create veth and set it UP")
	}
	// config ip address and route
	if err := moveVethToContainerAndSetIPAndRouteInContainerNetNs(endpoint, containerInitPid); err != nil {
		return nil, exception.NewGenericErrorWithContext(err, exception.SystemError, "set veth ip and route")
	}
	// config port mapping
	if err := setupPortMappings(endpoint); err != nil {
		return nil, exception.NewGenericErrorWithContext(err, exception.SystemError, "set up port mappings")
	}
	return endpoint, nil
}

func (driver *BridgeNetworkDriver) Disconnect(endpoint *Endpoint) error {
	// 回收IP地址
	allocator, err := LoadIPAllocator()
	if err != nil {
		return err
	}
	if err := allocator.Release(&endpoint.Network.IpRange, endpoint.IpAddress); err != nil {
		logrus.Warnf(err.Error())
	}
	// 删除端口映射
	if err := deletePortMappings(endpoint); err != nil {
		logrus.Warnf(err.Error())
	}
	// 删除网络端点
	hostVeth, err := netlink.LinkByName(endpoint.Device.Name)
	if err != nil {
		logrus.Warnf(err.Error())
	}
	if err := netlink.LinkDel(hostVeth); err != nil {
		logrus.Warnf(err.Error())
	}
	return nil
}
