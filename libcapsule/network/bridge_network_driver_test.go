package network

import "testing"

var driver = BridgeNetworkDriver{}

func TestBridgeNetworkDriver_Create(t *testing.T) {
}

func TestBridgeNetworkDriver_Load(t *testing.T) {

}

func TestBridgeNetworkDriver_Name(t *testing.T) {
	if driver.Name() != "bridge" {
		t.FailNow()
	}
}

func TestBridgeNetworkDriver_Delete(t *testing.T) {

}

func TestBridgeNetworkDriver_Connect(t *testing.T) {

}

func TestBridgeNetworkDriver_Disconnect(t *testing.T) {

}
