package network

import (
	"github.com/songxinjianqwe/capsule/libcapsule/constant"
	"net"
	"path/filepath"
	"sync"
)

// ipam is short for ip address management
type IPAM interface {
	Allocate(subnet *net.IPNet) (net.IP, error)
	Release(subnet *net.IPNet, ip net.IP) error
	Allocatable(subnet *net.IPNet) uint
}

var onceForIPAM sync.Once
var singletonIPAM *LocalIPAM
var singletonErr error

func LoadIPAllocator(runtimeRoot string) (IPAM, error) {
	onceForIPAM.Do(func() {
		singletonIPAM = &LocalIPAM{
			subnetAllocatorPath: filepath.Join(runtimeRoot, constant.IPAMDefaultAllocatorPath),
		}
		singletonErr = singletonIPAM.load()
	})
	return singletonIPAM, singletonErr
}
