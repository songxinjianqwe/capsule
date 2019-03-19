package network

import (
	"github.com/coreos/go-iptables/iptables"
	"github.com/stretchr/testify/assert"
	"github.com/vishvananda/netlink"
	"net"
	"strings"
	"testing"
)

var driver = BridgeNetworkDriver{}

func TestBridgeNetworkDriver_Create_Load_Delete(t *testing.T) {
	subnet := "192.168.10.0/24"
	name := "test_bridge0"
	defer driver.Delete(name)
	createdNetwork, err := driver.Create(subnet, name)
	assert.Nil(t, err)
	// 如果test失败也要把这个删掉
	network, err := driver.Load(name)
	assert.Nil(t, err)
	t.Logf("Network: %s", network)

	checkBridgeData(t, name, *createdNetwork.Subnet(), network)
	checkBridgeRoute(t, name, subnet)
	checkBridgeUp(t, name)
	checkSNAT(t, name, subnet)

	err = driver.Delete(name)
	assert.Nil(t, err)
	_, err = driver.Load(name)
	assert.NotNil(t, err, "delete network did not work")
}

func checkBridgeData(t *testing.T, name string, ipRange net.IPNet, network *Network) {
	assert.Equal(t, name, network.Name, "network name is wrong")
	assert.Equal(t, "bridge", network.Driver, "network driver is wrong")
	assert.Equal(t, ipRange.String(), network.Subnet().String(), "network addr is wrong")
}

func checkBridgeRoute(t *testing.T, name string, subnet string) {
	link, err := netlink.LinkByName(name)
	assert.Nil(t, err)
	routes, err := netlink.RouteList(link, netlink.FAMILY_ALL)
	assert.Nil(t, err)
	var routeExist bool
	for _, route := range routes {
		t.Logf("route: %s", route)
		if route.Dst.String() == subnet {
			routeExist = true
		}
	}
	assert.True(t, routeExist, "route not exists")
}

func checkBridgeUp(t *testing.T, name string) {
	link, err := netlink.LinkByName(name)
	assert.Nil(t, err)
	t.Logf("link: %s", link)
	t.Logf("link flag: %s", link.Attrs().Flags.String())
	assert.True(t, strings.Contains(link.Attrs().Flags.String(), "up"), "network is down")
}

func checkSNAT(t *testing.T, name string, subnet string) {
	ipRange, err := netlink.ParseIPNet(subnet)
	table, err := iptables.New()
	assert.Nil(t, err)
	rules, err := table.List("nat", "POSTROUTING")
	assert.Nil(t, err)
	for i, rule := range rules {
		t.Logf("[RULE %d]%s", i, rule)
	}
	exists, err := table.Exists(
		"nat",
		"POSTROUTING",
		getSNATRuleSpecs(name, *ipRange)...)
	assert.Nil(t, err)
	assert.True(t, exists, "SNAT Rule do not exist")
}

func TestBridgeNetworkDriver_Name(t *testing.T) {
	assert.Equal(t, "bridge", driver.Name())
}

func TestBridgeNetworkDriver_Connect(t *testing.T) {

}

func TestBridgeNetworkDriver_Disconnect(t *testing.T) {

}
