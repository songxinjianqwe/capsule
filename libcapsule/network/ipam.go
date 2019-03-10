package network

import (
	"github.com/songxinjianqwe/capsule/libcapsule"
	"net"
	"sync"
)

const IPAMDefaultAllocatorPath = libcapsule.RuntimeRoot + "/network/ipam/subnet.json"

// ipam is short for ip address management
type IPAM interface {
	Allocate(subnet *net.IPNet) (net.IP, error)
	Release(subnet *net.IPNet, ip *net.IP) error
}

var once sync.Once
var singletonIPAM *DefaultIPAMImpl
var singletonErr error

func LoadIPAllocator() (IPAM, error) {
	once.Do(func() {
		singletonIPAM = &DefaultIPAMImpl{
			subnetAllocatorPath: IPAMDefaultAllocatorPath,
		}
		singletonErr = singletonIPAM.load()
	})
	return singletonIPAM, singletonErr
}
