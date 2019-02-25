package util

import (
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule"
	specutil "github.com/songxinjianqwe/rune/libcapsule/util/spec"
	"os"
)

type ContainerAction uint8

const (
	ContainerActCreate ContainerAction = iota
	ContainerActRun
)

func (action ContainerAction) String() string {
	switch action {
	case ContainerActCreate:
		return "ContainerActCreate"
	case ContainerActRun:
		return "ContainerActRun"
	default:
		return "Unknown ContainerAction"
	}
}

/**
创建或启动容器
create
or
create and start
*/
func LaunchContainer(id string, spec *specs.Spec, action ContainerAction, init bool, detach bool) (int, error) {
	logrus.Infof("launching container:%s, action: %s", id, action)
	var container libcapsule.Container
	var err error
	if init {
		container, err = CreateContainer(id, spec)
	} else {
		container, err = GetContainer(id)
	}
	if err != nil {
		return -1, err
	}
	// 将specs.Process转为libcapsule.Process
	process, err := newProcess(*spec.Process, init, detach)
	logrus.Infof("new process complete, libcapsule.Process: %#v", process)
	if err != nil {
		return -1, err
	}
	switch action {
	case ContainerActCreate:
		err := container.Create(process)
		if err != nil {
			return -1, err
		}
		// 不管是否是terminal，都不会删除容器
	case ContainerActRun:
		// c.run == c.start + c.exec [+ c.destroy]
		err := container.Run(process)
		if err != nil {
			return -1, err
		}
		// 如果是前台运行，那么Run结束，即为容器进程结束，则删除容器
		if !detach {
			if err := container.Destroy(); err != nil {
				return -1, err
			}
		}
	}
	return 0, nil
}

/**
根据id读取一个Container
*/
func GetContainer(id string) (libcapsule.Container, error) {
	if id == "" {
		return nil, fmt.Errorf("container id cannot be empty")
	}
	factory, err := LoadFactory()
	if err != nil {
		return nil, err
	}
	return factory.Load(id)
}

/**
创建容器实例
*/
func CreateContainer(id string, spec *specs.Spec) (libcapsule.Container, error) {
	logrus.Infof("creating container: %s", id)
	if id == "" {
		return nil, fmt.Errorf("container id cannot be empty")
	}
	// 1、将spec转为容器config
	config, err := specutil.CreateContainerConfig(id, spec)
	logrus.Infof("convert complete, config: %#v", config)
	if err != nil {
		return nil, err
	}
	// 2、创建容器工厂
	factory, err := LoadFactory()
	if err != nil {
		return nil, err
	}
	// 3、创建容器
	container, err := factory.Create(id, config)
	if err != nil {
		return nil, err
	}
	return container, nil
}

/*
创建容器工厂
*/
func LoadFactory() (libcapsule.Factory, error) {
	factory, err := libcapsule.NewFactory()
	if err != nil {
		return nil, err
	}
	return factory, nil
}

/*
将specs.Process转为libcapsule.Process
*/
func newProcess(p specs.Process, init bool, detach bool) (*libcapsule.Process, error) {
	logrus.Infof("converting specs.Process to libcapsule.Process")
	libcapsuleProcess := &libcapsule.Process{
		Args:            p.Args,
		Env:             p.Env,
		User:            fmt.Sprintf("%d:%d", p.User.UID, p.User.GID),
		Cwd:             p.Cwd,
		NoNewPrivileges: &p.NoNewPrivileges,
		Init:            init,
		Detach:          detach,
	}
	for _, posixResourceLimit := range p.Rlimits {
		rl, err := specutil.CreateResourceLimit(posixResourceLimit)
		if err != nil {
			return nil, err
		}
		libcapsuleProcess.ResourceLimits = append(libcapsuleProcess.ResourceLimits, rl)
	}
	// 如果启用终端，则将进程的stdin等置为os的
	if detach {
		libcapsuleProcess.Stdin = os.Stdin
		libcapsuleProcess.Stdout = os.Stdout
		libcapsuleProcess.Stderr = os.Stderr
	}
	return libcapsuleProcess, nil
}
