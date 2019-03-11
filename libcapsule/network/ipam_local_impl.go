package network

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"path"
	"sync"
)

type LocalIPAM struct {
	subnetAllocatorPath string
	subnetMap           map[string]string
	mutex               sync.Mutex
}

func (ipam *LocalIPAM) Allocate(subnet *net.IPNet) (net.IP, error) {
	ipam.mutex.Lock()
	defer ipam.mutex.Unlock()
	panic("implement me")
}

func (ipam *LocalIPAM) Release(subnet *net.IPNet, ip *net.IP) error {
	ipam.mutex.Lock()
	defer ipam.mutex.Unlock()
	panic("implement me")
}

func (ipam *LocalIPAM) load() error {
	// load
	if _, err := os.Stat(singletonIPAM.subnetAllocatorPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	bytes, err := ioutil.ReadFile(singletonIPAM.subnetAllocatorPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(bytes, &ipam.subnetMap); err != nil {
		return err
	}
	return nil
}

func (ipam *LocalIPAM) dump() error {
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
