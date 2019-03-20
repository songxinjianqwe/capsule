package network

import (
	"os/user"
	"testing"
)

func TestMain(m *testing.M) {
	userObj, _ := user.Current()
	ipam, _ := NewMemoryIPAllocator()
	driver = BridgeNetworkDriver{runtimeRoot: userObj.HomeDir, allocator: ipam}

	allocator, _ = NewMemoryIPAllocator()
	m.Run()
}
