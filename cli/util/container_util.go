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
create and run
*/
func LaunchContainer(id string, spec *specs.Spec, action ContainerAction) (int, error) {
	logrus.Infof("launching container:%s, action: %s", id, action)
	container, err := CreateContainer(id, spec)
	if err != nil {
		return -1, err
	}
	// 将specs.Process转为libcapsule.Process
	process, err := newProcess(*spec.Process, true)
	logrus.Infof("new process complete, libcapsule.Process: %#v", process)
	if err != nil {
		return -1, err
	}
	switch action {
	case ContainerActCreate:
		err := container.Start(process)
		if err != nil {
			return -1, err
		}
	case ContainerActRun:
		// c.run == c.start + c.exec
		err := container.Run(process)
		if err != nil {
			return -1, err
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
func newProcess(p specs.Process, init bool) (*libcapsule.Process, error) {
	logrus.Infof("converting specs.Process to libcapsule.Process")
	libcapsuleProcess := &libcapsule.Process{
		Args:            p.Args,
		Env:             p.Env,
		User:            fmt.Sprintf("%d:%d", p.User.UID, p.User.GID),
		Cwd:             p.Cwd,
		Label:           p.SelinuxLabel,
		NoNewPrivileges: &p.NoNewPrivileges,
		Init:            init,
	}

	if p.ConsoleSize != nil {
		libcapsuleProcess.ConsoleWidth = uint16(p.ConsoleSize.Width)
		libcapsuleProcess.ConsoleHeight = uint16(p.ConsoleSize.Height)
	}
	for _, posixResourceLimit := range p.Rlimits {
		rl, err := specutil.CreateResourceLimit(posixResourceLimit)
		if err != nil {
			return nil, err
		}
		libcapsuleProcess.Rlimits = append(libcapsuleProcess.Rlimits, rl)
	}
	if p.Terminal {
		libcapsuleProcess.Stdin = os.Stdin
		libcapsuleProcess.Stdout = os.Stdout
		libcapsuleProcess.Stderr = os.Stderr
	}
	return libcapsuleProcess, nil
}
