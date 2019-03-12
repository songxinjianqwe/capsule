package network

import (
	"fmt"
	"github.com/coreos/go-iptables/iptables"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
	"net"
	"os"
	"runtime"
	"strings"
)

func createBridgeInterface(name string) error {
	if _, err := net.InterfaceByName(name); err == nil || !strings.Contains(err.Error(), "no such network interface") {
		return fmt.Errorf("brigde name %s exists", name)
	}
	linkAttrs := netlink.NewLinkAttrs()
	linkAttrs.Name = name
	br := &netlink.Bridge{
		LinkAttrs: linkAttrs,
	}
	if err := netlink.LinkAdd(br); err != nil {
		return err
	}
	return nil
}

// SNAT MASQUERADE
func setupIPTablesMasquerade(name string, subnet net.IPNet) error {
	logrus.Infof("setting up iptables masquerade for %s", name)
	tables, err := iptables.New()
	if err != nil {
		return err
	}
	// iptables -t nat -A POSTROUTING -s %s -o %s -j MASQUERADE
	if err := tables.Append(
		"nat",
		"POSTROUTING", getSNATRuleSpecs(name, subnet)...); err != nil {
		return err
	}
	return nil
}

func getSNATRuleSpecs(name string, subnet net.IPNet) []string {
	return []string{
		fmt.Sprintf("-s%s", subnet.String()),
		fmt.Sprintf("-o%s", name),
		"-jMASQUERADE",
	}
}

// 启用
func setInterfaceUp(name string) error {
	logrus.Infof("setting interface %s up", name)
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}
	// `ip link set $link up`
	return netlink.LinkSetUp(iface)
}

// 设置网络接口的IP和路由
func setInterfaceIPAndRoute(name string, subnet string) error {
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}
	_, ipRange, err := net.ParseCIDR(subnet)
	logrus.Infof("set interface %s ip %s and route %s...", name, ipRange.IP.String(), ipRange.String())
	// ip addr add xxx
	// 做了两件事：
	// 1、配置了网络接口的IP地址(IP)
	// 2、配置了路由表，将来自该网段的网络请求转发到这个网络接口上
	addr := &netlink.Addr{
		IPNet: ipRange}
	// `ip addr add $addr dev $link`
	return netlink.AddrAdd(iface, addr)
}

func setupPortMappings(endpoint *Endpoint) error {
	tables, err := iptables.New()
	if err != nil {
		return err
	}
	for _, mapping := range endpoint.PortMappings {
		split := strings.Split(mapping, ":")
		hostPort := split[0]
		containerPort := split[1]
		logrus.Infof("setting up %s port mapping %s:%s", endpoint.Name, hostPort, containerPort)
		if err := tables.Append(
			"nat",
			"PREROUTING",
			"-ptcp",
			"-mtcp",
			fmt.Sprintf("--dport%s", hostPort),
			"-jDNAT",
			fmt.Sprintf("--to-destination%s:%s", endpoint.IpAddress.String(), containerPort),
		); err != nil {
			return err
		}
	}
	return nil
}

func deletePortMappings(endpoint *Endpoint) error {
	tables, err := iptables.New()
	if err != nil {
		return err
	}
	for _, mapping := range endpoint.PortMappings {
		split := strings.Split(mapping, ":")
		hostPort := split[0]
		containerPort := split[1]
		if err := tables.Delete(
			"nat",
			"PREROUTING",
			"-ptcp",
			"-mtcp",
			fmt.Sprintf("--dport%s", hostPort),
			"-jDNAT",
			fmt.Sprintf("--to-destination%s:%s", endpoint.IpAddress.String(), containerPort),
		); err != nil {
			return err
		}
	}
	return nil
}

func moveVethToContainerAndSetIPAndRouteInContainerNetNs(endpoint *Endpoint, pid int) error {
	containerVeth, err := netlink.LinkByName(endpoint.Device.PeerName)
	if err != nil {
		return err
	}
	originNetNsHandle, netNsFileHandle, err := enterContainerNetNs(pid)
	if err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "enter container net ns")
	}
	defer leaveContainerNetNs(originNetNsHandle, netNsFileHandle)
	// 下面就进入容器网络了
	logrus.Infof("moving veth %s to container, detail: %#v...", containerVeth.Attrs().Name, containerVeth)
	// 1.把veth移动到net ns中
	if err := netlink.LinkSetNsFd(containerVeth, int(netNsFileHandle.Fd())); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "move veth to container net ns")
	}

	// 2. 配置IP地址与路由
	// 此时interface的IP地址为endpoint的地址,而网段是bridge的网段
	// 将来自该网段的网络请求转发到这个网络接口上
	interfaceIP := endpoint.Network.IpRange
	interfaceIP.IP = endpoint.IpAddress
	if err := setInterfaceIPAndRoute(endpoint.Name, interfaceIP.String()); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "set veth ip and route")
	}

	// 3.启用
	if err := setInterfaceUp(endpoint.Name); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "set container veth UP")
	}

	// 4.启用loopback
	if err := setInterfaceUp("lo"); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "set container loopback UP")
	}

	// 5. 设置容器内的外部请求均通过容器内的veth端点访问
	// route add -net 0.0.0.0/0 gw $(bridge IP) dev $(veth端点设置)
	_, defaultIpRange, _ := net.ParseCIDR("0.0.0.0/0")
	defaultRoute := &netlink.Route{
		LinkIndex: containerVeth.Attrs().Index,
		Gw:        endpoint.Network.IpRange.IP,
		Dst:       defaultIpRange,
	}
	if err := netlink.RouteAdd(defaultRoute); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "add default route")
	}
	return nil
}

func enterContainerNetNs(pid int) (netns.NsHandle, *os.File, error) {
	logrus.Infof("entering container %d network namespace...", pid)
	netNsFileHandle, err := os.OpenFile(fmt.Sprintf("/proc/%d/ns/net", pid), os.O_RDONLY, 0)
	if err != nil {
		return -1, nil, err
	}
	netNsFd := netNsFileHandle.Fd()
	runtime.LockOSThread()
	originNetNsHandle, err := netns.Get()
	if err != nil {
		return -1, nil, err
	}
	if err := netns.Set(netns.NsHandle(netNsFd)); err != nil {
		return -1, nil, err
	}
	return originNetNsHandle, netNsFileHandle, nil
}

func leaveContainerNetNs(originNetNsHandle netns.NsHandle, netNsFileHandle *os.File) {
	logrus.Infof("leaving container network namespace...")
	if err := netns.Set(originNetNsHandle); err != nil {
		logrus.Errorf("leaving container net ns failed, cause: %s", err.Error())
	}
	if err := originNetNsHandle.Close(); err != nil {
		logrus.Errorf("leaving container net ns failed, cause: %s", err.Error())
	}
	runtime.UnlockOSThread()
	if err := netNsFileHandle.Close(); err != nil {
		logrus.Errorf("leaving container net ns failed, cause: %s", err.Error())
	}
}

func createVethPairAndSetUp(endpoint *Endpoint) error {
	logrus.Infof("create veth pair %#v and set it up...", endpoint.Name)
	bridge, err := netlink.LinkByName(endpoint.Network.Name)
	if err != nil {
		return err
	}
	vethAttrs := netlink.NewLinkAttrs()
	// link名称长度有限制
	vethAttrs.Name = endpoint.Name[:5]
	// 将一端连接到bridge上
	vethAttrs.MasterIndex = bridge.Attrs().Index

	endpoint.Device = netlink.Veth{
		LinkAttrs: vethAttrs,
		PeerName:  fmt.Sprintf("cif-%s", endpoint.Name[:5]),
	}

	if err := netlink.LinkAdd(&endpoint.Device); err != nil {
		return err
	}
	logrus.Infof("veth pair created: %#v", endpoint.Device)
	if err := netlink.LinkSetUp(&endpoint.Device); err != nil {
		return err
	}
	return nil
}
