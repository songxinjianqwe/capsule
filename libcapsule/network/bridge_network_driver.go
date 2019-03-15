package network

import (
	"fmt"
	"github.com/coreos/go-iptables/iptables"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
	"github.com/vishvananda/netlink"
	"net"
	"strings"
)

type BridgeNetworkDriver struct {
}

func (driver *BridgeNetworkDriver) Name() string {
	return "bridge"
}

func (driver *BridgeNetworkDriver) NetworkLabel() string {
	return "capsule_bridge_label"
}

func (driver *BridgeNetworkDriver) Create(subnet string, bridgeName string) (*Network, error) {
	// 如果subnet的格式是192.168.1.2/24，那么parseCIDR的第一个返回值是IP地址,192.168.1.2，第二个返回值是IPNet类型，192.168.1.0/24
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
	logrus.Infof("allocated gateway ip: %s", gatewayIP.String())
	ipRange.IP = gatewayIP
	network := &Network{
		Name:    bridgeName,
		IpRange: *ipRange,
		Driver:  driver.Name(),
	}
	logrus.Infof("network: %s", network)

	// 1.创建bridge
	if err := createBridgeInterface(bridgeName, driver.NetworkLabel()); err != nil {
		return nil, exception.NewGenericErrorWithContext(err, exception.BridgeNetworkCreateError, "create bridge")
	}

	// 2.设置Bridge的IP地址和路由
	if err := setInterfaceIPAndRoute(bridgeName, *ipRange); err != nil {
		return nil, exception.NewGenericErrorWithContext(err, exception.InterfaceIPAndRouteSetError, "set bridge ip and route")
	}

	// 3.启动Bridge
	if err := setInterfaceUp(bridgeName); err != nil {
		return nil, exception.NewGenericErrorWithContext(err, exception.InterfaceSetUpError, "set bridge UP")
	}

	// 4.设置iptables SNAT规则（MASQUERADE）
	if err := setupIPTablesMasquerade(bridgeName, *ipRange); err != nil {
		return nil, exception.NewGenericErrorWithContext(err, exception.IPTablesSetError, "set iptables SNAT MASQUERADE RULE")
	}
	return network, nil
}

func (driver *BridgeNetworkDriver) Load(name string) (*Network, error) {
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return nil, exception.NewGenericError(err, exception.NetworkLinkNotFoundError)
	}
	//  `ip addr show`.
	addrs, err := netlink.AddrList(iface, netlink.FAMILY_ALL)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, exception.NewGenericError(fmt.Errorf("addresses not found"), exception.BridgeNetworkLoadError)
	}
	var bridgeAddr *net.IPNet
	for _, addr := range addrs {
		if addr.Label == name {
			bridgeAddr = addr.IPNet
			break
		}
	}
	if bridgeAddr == nil {
		return nil, exception.NewGenericError(fmt.Errorf("label-matched addresses not found"), exception.BridgeNetworkLoadError)
	}
	return &Network{
		Name:    name,
		Driver:  driver.Name(),
		IpRange: *bridgeAddr,
	}, nil
}

func (driver *BridgeNetworkDriver) List() ([]*Network, error) {
	links, err := netlink.LinkList()
	if err != nil {
		return nil, exception.NewGenericError(err, exception.NetworkLinkNotFoundError)
	}
	var networks []*Network
	for _, link := range links {
		if strings.HasPrefix(link.Attrs().Alias, driver.NetworkLabel()) {
			instance, err := driver.Load(link.Attrs().Name)
			if err != nil {
				return nil, err
			}
			networks = append(networks, instance)
		}
	}
	return networks, nil
}

func (driver *BridgeNetworkDriver) Delete(name string) error {
	network, err := driver.Load(name)
	if err != nil {
		return err
	}
	logrus.Infof("loaded network: %s", network)
	// 删除SNAT规则
	tables, err := iptables.New()
	if err := tables.Delete(
		"nat",
		"POSTROUTING",
		getSNATRuleSpecs(network.Name, network.IpRange)...,
	); err != nil {
		return exception.NewGenericError(err, exception.IPTablesDeleteError)
	}

	// 回收gateway IP
	allocator, err := LoadIPAllocator()
	if err != nil {
		return err
	}
	if err := allocator.Release(network.Subnet(), network.GatewayIP()); err != nil {
		return err
	}
	// 删除interface
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return exception.NewGenericError(err, exception.NetworkLinkNotFoundError)
	}
	if err := netlink.LinkDel(iface); err != nil {
		return exception.NewGenericError(err, exception.NetworkLinkDeleteError)
	}
	return nil
}

func (driver *BridgeNetworkDriver) Connect(endpointId string, network *Network, portMappings []string, containerInitPid int) (*Endpoint, error) {
	allocator, err := LoadIPAllocator()
	if err != nil {
		return nil, err
	}
	endpointIP, err := allocator.Allocate(network.Subnet())
	if err != nil {
		return nil, err
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
		return nil, exception.NewGenericErrorWithContext(err, exception.VethPairCreateError, "create veth and set it UP")
	}
	// config ip address and route
	if err := setUpContainerVethInNetNs(endpoint, containerInitPid); err != nil {
		return nil, exception.NewGenericErrorWithContext(err, exception.VethInitError, "set veth ip and route")
	}
	// config port mapping
	if err := setupPortMappings(endpoint); err != nil {
		return nil, exception.NewGenericErrorWithContext(err, exception.PortMappingsConfigError, "set up port mappings")
	}
	return endpoint, nil
}

func (driver *BridgeNetworkDriver) Disconnect(endpoint *Endpoint) error {
	// 回收IP地址
	allocator, err := LoadIPAllocator()
	if err != nil {
		logrus.Warnf(err.Error())
		return err
	}
	logrus.Infof("before releasing, allocatable ip: %d", allocator.Allocatable(endpoint.Network.Subnet()))
	if err := allocator.Release(endpoint.Network.Subnet(), endpoint.IpAddress); err != nil {
		logrus.Warnf(err.Error())
		return err
	}
	logrus.Infof("after releasing, allocatable ip: %d", allocator.Allocatable(endpoint.Network.Subnet()))
	// 删除端口映射
	if err := deletePortMappings(endpoint); err != nil {
		logrus.Warnf(err.Error())
	}
	// 删除宿主机上的网络端点(前面kill掉容器init process后,容器net namespace被销毁,容器内veth被销毁,宿主机与之peer的veth也随之被销毁)
	return err
}
