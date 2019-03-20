package network

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

var allocator IPAM

func TestLocalIPAM_Allocate_Release(t *testing.T) {
	_, subnet, _ := net.ParseCIDR("192.168.1.0/24")
	originalAllocatable := allocator.Allocatable(subnet)
	t.Logf("originalAllocatable: %d", originalAllocatable)
	ip, err := allocator.Allocate(subnet)
	assert.Nil(t, err)
	t.Logf("allocated IP: %s", ip)
	defer allocator.Release(subnet, ip)
	assert.Equal(t, uint(originalAllocatable-1), allocator.Allocatable(subnet))

	err = allocator.Release(subnet, ip)
	assert.Nil(t, err)
	assert.Equal(t, originalAllocatable, allocator.Allocatable(subnet))
}
