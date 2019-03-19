package libcapsule

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/network"
	"github.com/songxinjianqwe/capsule/libcapsule/util"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
	"syscall"
)

/*
init进程的启动hook
*/
func initStartHook(p *ParentAbstractProcess) error {
	// 将pid加入到cgroup set中
	if err := p.container.cgroupManager.JoinCgroupSet(p.pid()); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.CgroupsError, "applying cgroup configuration for process")
	}
	util.PrintSubsystemPids("memory", p.container.id, "after cgroup manager init", false)

	// 设置cgroup config
	if err := p.container.cgroupManager.SetConfig(p.container.config.Cgroup); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.CgroupsError, "setting cgroup config for procHooks process")
	}

	// 创建网络接口
	if err := createNetworkInterfaces(p); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.NetworkError, "creating network interfaces")
	}

	// init process会在启动后阻塞，直至收到config
	if err := p.sendConfigAndClosePipe(); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.PipeError, "sending config to init process")
	}

	// 等待init process到达在初始化之后，执行命令之前的状态
	// 使用SIGUSR1信号
	logrus.Info("start waiting init process ready(SIGUSR1) or fail(SIGCHLD) signal...")
	sig := util.WaitSignal(syscall.SIGUSR1, syscall.SIGCHLD)
	if sig == syscall.SIGUSR1 {
		logrus.Info("received SIGUSR1 signal")
	} else if sig == syscall.SIGCHLD {
		logrus.Errorf("received SIGCHLD signal")
		return fmt.Errorf("init process init failed")
	}
	return nil
}

// ******************************************************************************************************
// biz methods
// ******************************************************************************************************

func createNetworkInterfaces(p *ParentAbstractProcess) error {
	logrus.Infof("creating network interfaces")
	// 创建一个Bridge，如果没有的话
	var bridge *network.Network
	bridge, err := network.LoadNetwork("bridge", network.DefaultBridgeName)
	if err != nil {
		bridge, err = network.CreateNetwork("bridge", network.DefaultSubnet, network.DefaultBridgeName)
		if err != nil {
			return err
		}
	}
	logrus.Infof("create or load bridge complete, bridge: %#v", bridge)

	// 创建端点
	endpointConfig := p.container.config.Endpoint
	logrus.Infof("creating endpoint: %#v", endpointConfig)
	endpoint, err := network.Connect(endpointConfig.ID, endpointConfig.NetworkName, endpointConfig.PortMappings, p.pid())
	if err != nil {
		return err
	}
	p.container.endpoint = endpoint
	return nil
}
