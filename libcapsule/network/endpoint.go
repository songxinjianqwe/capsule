package network

import (
	"github.com/vishvananda/netlink"
	"net"
)

type Endpoint struct {
	ID          string
	Device      netlink.Veth
	IpAddress   net.IP
	MacAddress  net.HardwareAddr
	PortMapping []string
	Network     *Network
}
