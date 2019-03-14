package network

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
	"github.com/willf/bitset"
	"io/ioutil"
	"net"
	"os"
	"path"
	"sync"
)

// BitSet的文档:
// https://godoc.org/github.com/willf/bitset#BitSet
type LocalIPAM struct {
	subnetAllocatorPath string
	subnetMap           map[string]*bitset.BitSet
	mutex               sync.Mutex
}

func (ipam *LocalIPAM) Allocatable(subnet *net.IPNet) uint {
	ipam.mutex.Lock()
	defer ipam.mutex.Unlock()
	total := allocatableIPAmount(subnet)
	if _, exist := ipam.subnetMap[subnet.String()]; !exist {
		logrus.Infof("subnet %s not found, return full", subnet)
		return total
	}
	bitmap := ipam.subnetMap[subnet.String()]
	return total - bitmap.Count()
}

func (ipam *LocalIPAM) Allocate(subnet *net.IPNet) (net.IP, error) {
	ipam.mutex.Lock()
	defer ipam.mutex.Unlock()
	logrus.Infof("allocating ip in subnet:%s", subnet)
	if _, exist := ipam.subnetMap[subnet.String()]; !exist {
		amount := allocatableIPAmount(subnet)
		logrus.Infof("subnet %s do not exist, allocatable ip amount is %d", subnet, amount)
		ipam.subnetMap[subnet.String()] = bitset.New(amount)
	}
	bitmap := ipam.subnetMap[subnet.String()]
	logrus.Infof("bitmap: %s", bitmap)
	nextClearIndex, allocatable := bitmap.NextClear(0)
	if !allocatable {
		// 说明全部为1,则
		return nil, exception.NewGenericError(fmt.Errorf("no allocatable ip"), exception.IPRunOutError)
	}
	// gotcha!
	bitmap.Set(nextClearIndex)
	// 网段IP,注意这里一定要拷贝!!!
	ip := make([]byte, 4)
	copy(ip, subnet.IP.To4())
	// 假设subnet为192.168.1.0/24, index为184,那么IP地址为192.168.1.1+184=192.168.1.185
	// 184 = 0000 0000 0000 0000 0000 0000 1011 1000
	// loop0: ip[0] += uint8(0000 0000 0000 0000 0000 0000 1011 1000 >> 24)(即0000)
	// loop1: ip[1] += uint8(0000 0000 0000 0000 0000 0000 1011 1000 >> 16)(即0000)
	// loop2: ip[2] += uint8(0000 0000 0000 0000 0000 0000 1011 1000 >>  8)(即0000)
	// loop3: ip[3] += uint8(0000 0000 0000 0000 0000 0000 1011 1000 >>  0)(即1011 1000,184)
	for byteIndex := 4; byteIndex > 0; byteIndex-- {
		ip[4-byteIndex] += uint8(nextClearIndex >> uint8((byteIndex-1)*8))
	}
	// ip 从1开始
	ip[3]++
	logrus.Infof("allocated ip: %s", net.IP(ip).String())
	if err := ipam.dump(); err != nil {
		return nil, err
	}
	return ip, nil
}

func (ipam *LocalIPAM) Release(subnet *net.IPNet, ip net.IP) error {
	ipam.mutex.Lock()
	defer ipam.mutex.Unlock()
	if _, exist := ipam.subnetMap[subnet.String()]; !exist {
		return exception.NewGenericError(fmt.Errorf("subnet %s not exists", subnet), exception.IPReleaseError)
	}
	logrus.Infof("releasing ip %s in subnet:%s", ip, subnet)
	releasingIP := ip.To4()
	releasingIP[3]--
	var index uint
	// 假设subnet为192.168.1.0/24, IP地址为192.168.1.185
	// releasingIP 此时为[192, 168, 1, 184]
	// loop0: index += (184 - 0) << 0 -> index = 184
	// loop1: index += (1 - 1) << 8 -> index = 184
	// loop2: index += (168 - 168) << 16 -> index = 184
	// loop3: index += (192 - 192) << 24 -> index = 184
	for byteIndex := 4; byteIndex > 0; byteIndex-- {
		index += uint(releasingIP[byteIndex-1]-subnet.IP[byteIndex-1]) << uint((4-byteIndex)*8)
	}
	bitmap := ipam.subnetMap[subnet.String()]
	bitmap.Clear(index)
	if err := ipam.dump(); err != nil {
		return err
	}
	return nil
}

func (ipam *LocalIPAM) load() error {
	// load
	if _, err := os.Stat(singletonIPAM.subnetAllocatorPath); err != nil {
		if os.IsNotExist(err) {
			// 不存在，则构造一个新的Map
			ipam.subnetMap = make(map[string]*bitset.BitSet)
			return nil
		} else {
			return exception.NewGenericError(err, exception.IPAMLoadError)
		}
	}
	bytes, err := ioutil.ReadFile(singletonIPAM.subnetAllocatorPath)
	if err != nil {
		return exception.NewGenericError(err, exception.IPAMLoadError)
	}
	logrus.Infof("loaded subnetMap:%v", string(bytes))
	if err := json.Unmarshal(bytes, &ipam.subnetMap); err != nil {
		return exception.NewGenericError(err, exception.IPAMLoadError)
	}
	return nil
}

func (ipam *LocalIPAM) dump() error {
	if _, err := os.Stat(ipam.subnetAllocatorPath); err != nil && os.IsNotExist(err) {
		// 如果文件之前不存在，则先创建目录，再创建文件
		// 否则覆盖原来的文件
		dir := path.Dir(ipam.subnetAllocatorPath)
		if err := os.MkdirAll(dir, 0644); err != nil {
			return exception.NewGenericError(err, exception.IPAMDumpError)
		}
	}
	subnetFile, err := os.OpenFile(ipam.subnetAllocatorPath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return exception.NewGenericError(err, exception.IPAMDumpError)
	}
	defer subnetFile.Close()
	bytes, err := json.Marshal(ipam.subnetMap)
	if err != nil {
		return exception.NewGenericError(err, exception.IPAMDumpError)
	}
	if _, err := subnetFile.Write(bytes); err != nil {
		return exception.NewGenericError(err, exception.IPAMDumpError)
	}
	return nil
}

func allocatableIPAmount(subnet *net.IPNet) uint {
	// IP地址是32位，有子网情况下是 网段:子网，前面n位是网段地址，后面32-n是子网地址
	// subnet如果是192.168.1.0/24，那么子网掩码为255.255.255.0
	// ones为/24中的24位，bits为总共位数，其实就是32
	// 那么可分配的IP地址数量为2^(bits - one)=2^8=256个
	netSegmentBits, totalBits := subnet.Mask.Size()
	return uint(1 << uint8(totalBits-netSegmentBits))
}
