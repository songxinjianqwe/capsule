package network

import (
	"github.com/songxinjianqwe/capsule/libcapsule/constant"
	"github.com/willf/bitset"
	"net"
	"path/filepath"
)

// ipam is short for ip address management
type IPAM interface {
	Allocate(subnet *net.IPNet) (net.IP, error)
	Release(subnet *net.IPNet, ip net.IP) error
	Allocatable(subnet *net.IPNet) uint
}

func NewPersistentIPAllocator(runtimeRoot string) (IPAM, error) {
	ipam := &LocalIPAM{
		subnetAllocatorPath: filepath.Join(runtimeRoot, constant.IPAMDefaultAllocatorPath),
		mode:                IPAMPersistentMode,
	}
	if err := ipam.load(); err != nil {
		return nil, err
	}
	return ipam, nil
}

func NewMemoryIPAllocator() (IPAM, error) {
	ipam := &LocalIPAM{
		subnetMap: make(map[string]*bitset.BitSet),
		mode:      IPAMMemoryMode,
	}
	return ipam, nil
}
