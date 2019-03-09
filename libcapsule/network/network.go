package network

import "net"

type Network struct {
	// 网络名称
	Name string
	// 网段
	IpRange *net.IPNet
	// 网络驱动名（网络类型）
	Driver string
}
