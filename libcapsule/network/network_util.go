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

func createBridgeInterface(name string, networkLabel string) error {
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
	if err := netlink.LinkSetAlias(br, networkLabel+"-"+name); err != nil {
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
	_, ipNet, _ := net.ParseCIDR(subnet.String())
	// !的意思是negative,out设备名是除了name之外的其他网络设备
	// SNAT转换时将源IP转为某设备名,这里我们不清楚有哪些网卡,于是我们将其设置为除了我们bridge外的设备
	// 因为bridge的IP也是私有IP,外网是不认识的
	return []string{
		fmt.Sprintf("-s%s", ipNet.String()),
		"!",
		fmt.Sprintf("-o%s", name),
		"-jMASQUERADE",
	}
}

func getDNATRuleSpecs(containerIP string, hostPort string, containerPort string) []string {
	return []string{"-ptcp",
		"-mtcp",
		"-jDNAT",
		"--dport",
		hostPort,
		"--to-destination",
		fmt.Sprintf("%s:%s", containerIP, containerPort)}
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
// subnet可以同时存储IP地址和网段地址
func setInterfaceIPAndRoute(name string, interfaceIPAndRoute net.IPNet) error {
	ip, ipRange, _ := net.ParseCIDR(interfaceIPAndRoute.String())
	logrus.Infof("set interface %s ip %s and route %s", name, ip, ipRange)
	iface, err := netlink.LinkByName(name)
	if err != nil {
		return err
	}
	// ipRange包含两个信息
	if err != nil {
		return err
	}
	// ip addr add xxx
	// 做了两件事：
	// 1、配置了网络接口的IP地址(IP)
	// 2、配置了路由表，将来自该网段的网络请求转发到这个网络接口上
	// 设置Broadcast为nil,即为0.0.0.0,不能设置为网段的广播地址,否则会出现ARP找不到容器内IP的情况

	addr := &netlink.Addr{IPNet: &interfaceIPAndRoute, Peer: &interfaceIPAndRoute, Label: "", Flags: 0, Scope: 0}
	// `ip addr add $addr dev $link`
	if err := netlink.AddrAdd(iface, addr); err != nil {
		logrus.Errorf("config ip and route failed, cause: %s", err.Error())
		return err
	}
	return nil
}

// DNAT
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
			getDNATRuleSpecs(endpoint.IpAddress.String(), hostPort, containerPort)...,
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
			getDNATRuleSpecs(endpoint.IpAddress.String(), hostPort, containerPort)...,
		); err != nil {
			return err
		}
	}
	return nil
}

func setUpContainerVethInNetNs(endpoint *Endpoint, pid int) error {
	containerVethName := endpoint.GetContainerVethName()
	containerVeth, err := netlink.LinkByName(containerVethName)
	if err != nil {
		return err
	}
	netNsFileHandle, err := os.OpenFile(fmt.Sprintf("/proc/%d/ns/net", pid), os.O_RDONLY, 0)
	if err != nil {
		return err
	}

	// 1.把veth移动到net ns中
	if err := netlink.LinkSetNsFd(containerVeth, int(netNsFileHandle.Fd())); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "move veth to container net ns")
	}

	// 2. 进入container network namespace
	originNetNsHandle, err := enterContainerNetNs(int(netNsFileHandle.Fd()), pid)
	if err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "enter container net ns")
	}
	defer leaveContainerNetNs(originNetNsHandle, netNsFileHandle)
	// 下面就进入容器网络了
	logrus.Infof("moving veth %s to container, detail: %#v...", containerVeth.Attrs().Name, containerVeth)

	// 3. 配置IP地址与路由
	// 此时interface的IP地址为endpoint的地址,而网段是bridge的网段
	// 将来自该网段的网络请求转发到这个网络接口上
	interfaceIP := endpoint.Network.IpRange
	interfaceIP.IP = endpoint.IpAddress
	if err := setInterfaceIPAndRoute(containerVethName, interfaceIP); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "set veth ip and route")
	}

	// 4.启用
	if err := setInterfaceUp(containerVethName); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "set container veth UP")
	}

	// 5.启用loopback
	if err := setInterfaceUp("lo"); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "set container loopback UP")
	}

	// 6. 设置容器内的对外部的请求均通过容器内的veth端点访问
	// route add -net 0.0.0.0/0 gw $(bridge IP) dev $(veth端点设置)
	_, defaultIpRange, _ := net.ParseCIDR("0.0.0.0/0")
	defaultRoute := &netlink.Route{
		LinkIndex: containerVeth.Attrs().Index,
		Gw:        endpoint.Network.IpRange.IP,
		Dst:       defaultIpRange,
	}
	logrus.Infof("add default route in container: %s", defaultIpRange)
	if err := netlink.RouteAdd(defaultRoute); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SystemError, "add default route")
	}
	// 7.离开container network namespace
	return nil
}

func enterContainerNetNs(netNsFd int, pid int) (netns.NsHandle, error) {
	logrus.Infof("entering container %d network namespace...", pid)

	runtime.LockOSThread()
	originNetNsHandle, err := netns.Get()
	if err != nil {
		return -1, err
	}
	if err := netns.Set(netns.NsHandle(netNsFd)); err != nil {
		return -1, err
	}
	return originNetNsHandle, nil
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

	endpoint.Device = &netlink.Veth{
		LinkAttrs: vethAttrs,
		PeerName:  fmt.Sprintf("cif-%s", endpoint.Name[:5]),
	}
	if err := netlink.LinkAdd(endpoint.Device); err != nil {
		return err
	}
	logrus.Infof("veth pair created: %#v", endpoint.Device)
	if err := netlink.LinkSetUp(endpoint.Device); err != nil {
		return err
	}
	return nil
}
