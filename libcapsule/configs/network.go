package configs

import (
	"github.com/vishvananda/netlink"
	"net"
)

/*
对应一个网段，Driver取值有Bridge
*/
type Network struct {
	// 网络名称
	Name string
	// 网段
	IpRange *net.IPNet
	// 网络驱动名（网络类型）
	Driver string
}

/*
对应一个网络端点，比如容器中会有一个veth和一个loopback
*/
type Endpoint struct {
	ID string
	// veth 或者 loopback
	Type       string
	IpAddress  net.IP
	MacAddress net.HardwareAddr
	// 下面是Veth特有配置
	Device      netlink.Veth
	Network     *Network
	PortMapping []string
}
