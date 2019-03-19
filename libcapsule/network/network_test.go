package network

import (
	"os/user"
	"testing"
)

func TestMain(m *testing.M) {
	userObj, _ := user.Current()
	driver = BridgeNetworkDriver{runtimeRoot: userObj.HomeDir}
	allocator, _ = LoadIPAllocator(userObj.HomeDir)
	m.Run()
}
