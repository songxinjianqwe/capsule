package network

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
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
	// 如果不存在这个subnet，则用0填充
	if _, exist := ipam.subnetMap[subnet.String()]; !exist {
		// IP地址是32位，有子网情况下是 网段:子网，前面n位是网段地址，后面32-n是子网地址
		// subnet如果是192.168.1.0/24，那么子网掩码为255.255.255.0
		// ones为/24中的24位，bits为总共位数，其实就是32
		// 那么可分配的IP地址数量为2^(bits - one)=2^8=256个
		netSegmentBits, totalBits := subnet.Mask.Size()
		allocatableIPAmount := 1 << uint8(totalBits-netSegmentBits)
		logrus.Infof("subnet %s do not exist, allocatable ip amount is %d", subnet, allocatableIPAmount)
		ipam.subnetMap[subnet.String()] = strings.Repeat("0", allocatableIPAmount)
	}
	return nil, nil
}

func (ipam *LocalIPAM) Release(subnet *net.IPNet, ip *net.IP) error {
	ipam.mutex.Lock()
	defer ipam.mutex.Unlock()
	panic("implement me")
}

func (ipam *LocalIPAM) load() error {
	// load
	if _, err := os.Stat(singletonIPAM.subnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			// 不存在，则构造一个新的Map
			ipam.subnetMap = make(map[string]string)
			return nil
		} else {
			return err
		}
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
