package util

import (
	"fmt"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule"
	"github.com/songxinjianqwe/capsule/libcapsule/constant"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
	specutil "github.com/songxinjianqwe/capsule/libcapsule/util/spec"
	"io/ioutil"
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

/*
进入容器执行一个Process
*/
func ExecContainer(id string, detach bool, args []string, cwd string, env []string) (string, error) {
	logrus.Infof("exec container: %s, detach: %t, args: %v, cwd: %s, env: %v", id, detach, args, cwd, env)
	container, err := GetContainer(id)
	if err != nil {
		return "", err
	}
	containerStatus, err := container.Status()
	if err != nil {
		return "", err
	}
	// exec时,先检查容器状态是否为Stopped
	if containerStatus == libcapsule.Stopped {
		return "", fmt.Errorf("cant exec in a stopped container ")
	}
	execId, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	ociState, err := container.OCIState()
	if err != nil {
		return "", err
	}
	spec, err := LoadSpec(ociState.Bundle)
	if err != nil {
		return "", err
	}
	// 构造一个Process，由命令行输入的参数会覆盖spec中的Init Process Config
	process, err := newProcess(execId.String(), spec.Process, false, detach)
	if err != nil {
		return "", err
	}
	// override
	process.Args = args
	if cwd != "" {
		process.Cwd = cwd
	}
	process.Env = append(process.Env, env...)

	logrus.Infof("new exec process complete, libcapsule.Process: %#v", process)
	// 无论是否是daemon运行，在执行完exec process后，都不会销毁容器。
	return execId.String(), container.Run(process)
}

/*
创建或启动容器
create
or
create and start
Process一定为Init Process
*/
func CreateOrRunContainer(id string, bundle string, spec *specs.Spec, action ContainerAction, detach bool, portMappings []string) error {
	logrus.Infof("create or run container: %s, action: %s", id, action)
	container, err := CreateContainer(id, bundle, spec, portMappings)
	if err != nil {
		return err
	}
	// 将specs.Process转为libcapsule.Process
	process, err := newProcess(id, spec.Process, true, detach)
	logrus.Infof("new init process complete, libcapsule.Process: %#v", process)
	if err != nil {
		return err
	}
	var containerErr error
	switch action {
	case ContainerActCreate:
		// 如果是create，那么不管是否是terminal，都不会删除容器
		containerErr = container.Create(process)
	case ContainerActRun:
		// c.run == c.start + c.exec [+ c.destroy]
		containerErr = container.Run(process)
	}
	if containerErr != nil {
		return handleContainerErr(container, containerErr)
	}
	// 如果是Run命令运行容器吗，并且是前台运行，那么Run结束，即为容器进程结束，则删除容器
	if action == ContainerActRun && !detach {
		if err := container.Destroy(); err != nil {
			return err
		}
	}
	return nil
}

func handleContainerErr(container libcapsule.Container, containerErr error) error {
	if err := container.Destroy(); err != nil {
		logrus.Errorf(fmt.Sprintf("container create failed, try to destroy it but failed again, cause: %s", containerErr.Error()))
		return exception.NewGenericErrorWithContext(
			err,
			exception.SystemError,
			fmt.Sprintf("container create failed, try to destroy it but failed again, cause: %s", containerErr.Error()))
	}
	return containerErr
}

/*
根据id读取一个Container
*/
func GetContainer(id string) (libcapsule.Container, error) {
	if id == "" {
		return nil, fmt.Errorf("container id cannot be empty")
	}
	factory, err := libcapsule.NewFactory(true)
	if err != nil {
		return nil, err
	}
	return factory.Load(id)
}

/*
查询所有的id
*/
func GetContainerIds() ([]string, error) {
	var ids []string
	if _, err := os.Stat(constant.ContainerRuntimeRoot); err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
	}
	list, err := ioutil.ReadDir(constant.ContainerRuntimeRoot)
	if err != nil {
		return nil, err
	}
	for _, fileInfo := range list {
		ids = append(ids, fileInfo.Name())
	}
	return ids, nil
}

/*
创建容器实例
*/
func CreateContainer(id string, bundle string, spec *specs.Spec, portMappings []string) (libcapsule.Container, error) {
	logrus.Infof("creating container: %s", id)
	if id == "" {
		return nil, fmt.Errorf("container id cannot be empty")
	}
	// 1、将spec转为容器config
	config, err := specutil.CreateContainerConfig(bundle, spec, portMappings)
	logrus.Infof("convert complete, config: %#v", config)
	if err != nil {
		return nil, err
	}
	// 2、创建容器工厂
	factory, err := libcapsule.NewFactory(true)
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
将specs.Process转为libcapsule.Process
*/
func newProcess(id string, p *specs.Process, init, detach bool) (*libcapsule.Process, error) {
	logrus.Infof("converting specs.Process to libcapsule.Process")
	libcapsuleProcess := &libcapsule.Process{
		ID:     id,
		Args:   p.Args,
		Env:    p.Env,
		Cwd:    p.Cwd,
		Init:   init,
		Detach: detach,
	}
	return libcapsuleProcess, nil
}
