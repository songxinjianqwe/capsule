package network

import (
	"encoding/json"
	"github.com/songxinjianqwe/capsule/libcapsule"
	"io/ioutil"
	"net"
	"os"
	"path"
	"sync"
)

const IPAMDefaultAllocatorPath = libcapsule.RuntimeRoot + "/network/ipam/subnet.json"

var once sync.Once
var singletonIPAM *DefaultIPAM
var singletonErr error

func LoadIPAllocator() (IPAM, error) {
	once.Do(func() {
		singletonIPAM = &DefaultIPAM{
			subnetAllocatorPath: IPAMDefaultAllocatorPath,
		}
		// load
		if _, err := os.Stat(singletonIPAM.subnetAllocatorPath); err != nil && !os.IsNotExist(err) {
			singletonErr = err
		}
		bytes, err := ioutil.ReadFile(singletonIPAM.subnetAllocatorPath)
		if err != nil {
			singletonErr = err
		}
		if err := json.Unmarshal(bytes, &singletonIPAM.subnetMap); err != nil {
			singletonErr = err
		}
	})
	return singletonIPAM, singletonErr
}

type IPAM interface {
	Allocate(subnet *net.IPNet) (net.IP, error)
	Release(subnet *net.IPNet, ip *net.IP) error
}

// ipam is short for ip address management
type DefaultIPAM struct {
	subnetAllocatorPath string
	subnetMap           map[string]string
}

func (ipam *DefaultIPAM) Allocate(subnet *net.IPNet) (net.IP, error) {
	panic("implement me")
}

func (ipam *DefaultIPAM) Release(subnet *net.IPNet, ip *net.IP) error {
	panic("implement me")
}

func (ipam *DefaultIPAM) dump() error {
	if _, err := os.Stat(ipam.subnetAllocatorPath); err != nil && os.IsNotExist(err) {
		// 如果文件之前不存在，则先创建目录，再创建文件
		// 否则覆盖原来的文件
		dir := path.Dir(ipam.subnetAllocatorPath)
		if err := os.MkdirAll(dir, 0644); err != nil {
			return err
		}
	}
	subnetFile, err := os.OpenFile(ipam.subnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer subnetFile.Close()
	bytes, err := json.Marshal(ipam.subnetMap)
	if err != nil {
		return err
	}
	if _, err := subnetFile.Write(bytes); err != nil {
		return err
	}
	return nil
}
