package network

import "testing"

var driver = BridgeNetworkDriver{}

func TestBridgeNetworkDriver_Create_Load_Delete(t *testing.T) {
	subnet := "192.168.10.1/24"
	name := "test_bridge0"
	_, err := driver.Create(subnet, name)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
	network, err := driver.Load(name)
	if err != nil {
		t.Errorf(err.Error())
		t.FailNow()
	}
	if network.Name != name {
		t.Errorf("network name is wrong: %s", network.Name)
		t.FailNow()
	}
	if network.Driver != "bridge" {
		t.Errorf("network driver is wrong: %s", network.Driver)
		t.FailNow()
	}
	t.Logf("Network: %s", network)
	if err := driver.Delete(name); err != nil {
		t.Errorf("delete network failed, cause: %s", err.Error())
		t.FailNow()
	}
	if _, err := driver.Load("docker0"); err == nil {
		t.Errorf("delete network failed")
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
