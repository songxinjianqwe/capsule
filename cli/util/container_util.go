package util

import (
	"errors"
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/songxinjianqwe/rune/libcapsule"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	specutil "github.com/songxinjianqwe/rune/libcapsule/util/spec"
	"path/filepath"
	"strconv"
)

var errEmptyID = errors.New("container id cannot be empty")

type ContainerAction uint8

const (
	ContainerActCreate ContainerAction = iota + 1
	ContainerActRun
)

/**
创建或启动容器
create
or
create and run
*/
func LaunchContainer(id string, spec *specs.Spec, action ContainerAction) (int, error) {
	container, err := CreateContainer(id, spec)
	if err != nil {
		return -1, err
	}
	// 将specs.Process转为libcapsule.Process
	process, err := newProcess(*spec.Process)
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
		return nil, errEmptyID
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
	// 将 process.cwd + rootfs 拼接作为Rootfs的路径
	if id == "" {
		return nil, errEmptyID
	}
	rootfsPath := spec.Root.Path
	if !filepath.IsAbs(rootfsPath) {
		rootfsPath = filepath.Join(spec.Process.Cwd, rootfsPath)
	}
	config, err := specutil.CreateContainerConfig(id, spec)
	if err != nil {
		return nil, err
	}
	factory, err := LoadFactory()
	if err != nil {
		return nil, err
	}
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
func newProcess(p specs.Process) (*libcapsule.Process, error) {
	lp := &libcapsule.Process{
		Args:            p.Args,
		Env:             p.Env,
		User:            fmt.Sprintf("%d:%d", p.User.UID, p.User.GID),
		Cwd:             p.Cwd,
		Label:           p.SelinuxLabel,
		NoNewPrivileges: &p.NoNewPrivileges,
	}

	if p.ConsoleSize != nil {
		lp.ConsoleWidth = uint16(p.ConsoleSize.Width)
		lp.ConsoleHeight = uint16(p.ConsoleSize.Height)
	}

	if p.Capabilities != nil {
		lp.Capabilities = &specs.LinuxCapabilities{}
		lp.Capabilities.Bounding = p.Capabilities.Bounding
		lp.Capabilities.Effective = p.Capabilities.Effective
		lp.Capabilities.Inheritable = p.Capabilities.Inheritable
		lp.Capabilities.Permitted = p.Capabilities.Permitted
		lp.Capabilities.Ambient = p.Capabilities.Ambient
	}
	for _, gid := range p.User.AdditionalGids {
		lp.AdditionalGroups = append(lp.AdditionalGroups, strconv.FormatUint(uint64(gid), 10))
	}
	for _, rlimit := range p.Rlimits {
		rl, err := createLibCapsuleRlimit(rlimit)
		if err != nil {
			return nil, err
		}
		lp.Rlimits = append(lp.Rlimits, rl)
	}
	return lp, nil
}

func createLibCapsuleRlimit(rlimit specs.POSIXRlimit) (configc.Rlimit, error) {
	rl, err := strToRlimit(rlimit.Type)
	if err != nil {
		return configc.Rlimit{}, err
	}
	return configc.Rlimit{
		Type: rl,
		Hard: rlimit.Hard,
		Soft: rlimit.Soft,
	}, nil
}
