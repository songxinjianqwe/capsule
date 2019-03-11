package network

import (
	"github.com/coreos/go-iptables/iptables"
	"github.com/vishvananda/netlink"
	"strings"
	"testing"
)

var driver = BridgeNetworkDriver{}

func TestBridgeNetworkDriver_Create_Load_Delete(t *testing.T) {
	subnet := "192.168.10.0/24"
	//ip, _, _ := net.ParseCIDR(subnet)
	name := "test_bridge0"
	_, err := driver.Create(subnet, name)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
	// 如果test失败也要把这个删掉
	defer driver.Delete(name)
	network, err := driver.Load(name)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
	t.Logf("Network: %s", network)
	// 检查network是否OK

	checkNetworkCorrect(t, name, subnet, network)
	checkBridgeRoute(t, name, subnet)
	checkBridgeUp(t, name)
	checkSNAT(t, name, subnet)

	if err := driver.Delete(name); err != nil {
		t.Errorf("delete network failed, cause: %s", err.Error())
		t.FailNow()
	}
	if _, err := driver.Load(name); err == nil {
		t.Errorf("delete network failed")
		t.FailNow()
	}
}

func checkNetworkCorrect(t *testing.T, name string, subnet string, network *Network) {
	if network.Name != name {
		t.Errorf("network name is wrong: %s", network.Name)
		t.FailNow()
	}
	if network.Driver != "bridge" {
		t.Errorf("network driver is wrong: %s", network.Driver)
		t.FailNow()
	}
	if network.IpRange.String() != subnet {
		t.Errorf("network subnet is wrong: %s", network.IpRange)
		t.FailNow()
	}
}

func checkBridgeRoute(t *testing.T, name string, subnet string) {
	link, err := netlink.LinkByName(name)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
	routes, err := netlink.RouteList(link, netlink.FAMILY_ALL)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
	var routeExist bool
	for _, route := range routes {
		t.Logf("route: %s", route)
		if route.Dst.String() == subnet {
			routeExist = true
		}
	}
	if !routeExist {
		t.Error("route not exists")
		t.FailNow()
	}
}

func checkBridgeUp(t *testing.T, name string) {
	link, err := netlink.LinkByName(name)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
	t.Logf("link: %s", link)
	t.Logf("link flag: %s", link.Attrs().Flags.String())
	if !strings.Contains(link.Attrs().Flags.String(), "up") {
		t.Errorf("network is down")
		t.FailNow()
	}
}

func checkSNAT(t *testing.T, name string, subnet string) {
	ipRange, err := netlink.ParseIPNet(subnet)
	table, err := iptables.New()
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
	exists, err := table.Exists(
		"nat",
		"POSTROUTING",
		getSNATRuleSpec(name, ipRange)...)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
	if !exists {
		t.Errorf("SNAT Rule do not exist")
		t.FailNow()
	}
}

func TestBridgeNetworkDriver_Name(t *testing.T) {
	if driver.Name() != "bridge" {
		t.FailNow()
	}
}

func TestBridgeNetworkDriver_Connect(t *testing.T) {

}

func TestBridgeNetworkDriver_Disconnect(t *testing.T) {

}
