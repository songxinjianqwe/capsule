package network

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

var allocator IPAM

func TestMain(m *testing.M) {
	allocator, _ = LoadIPAllocator()
	m.Run()
}

func TestLocalIPAM_Allocate(t *testing.T) {
	_, ipNet, _ := net.ParseCIDR("192.168.1.0/24")
	ip, err := allocator.Allocate(ipNet)
	assert.Nil(t, err)
	t.Logf("allocated IP: %s", ip)
}
