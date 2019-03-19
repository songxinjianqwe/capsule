package network

import (
	"encoding/json"
	"github.com/songxinjianqwe/capsule/libcapsule/constant"
	"github.com/stretchr/testify/assert"
	"github.com/willf/bitset"
	"io/ioutil"
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

	bytes, err := ioutil.ReadFile(constant.IPAMDefaultAllocatorPath)
	assert.Nil(t, err)
	subnetMap := make(map[string]*bitset.BitSet)
	err = json.Unmarshal(bytes, &subnetMap)
	assert.Nil(t, err)
	_, exist := subnetMap[subnet.String()]
	assert.True(t, exist)
	bitmap := subnetMap[subnet.String()]
	// bitmap中一定有1
	assert.True(t, bitmap.Any())

	err = allocator.Release(subnet, ip)
	assert.Nil(t, err)
	assert.Equal(t, originalAllocatable, allocator.Allocatable(subnet))
}
