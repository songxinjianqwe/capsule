package util

import (
	"github.com/songxinjianqwe/capsule/libcapsule"
	"github.com/songxinjianqwe/capsule/libcapsule/util"
	"time"
)

// state represents the platform agnostic pieces relating to a
// running container's status and state
type ContainerStateVO struct {
	// Version is the OCI version for the container
	Version string `json:"ociVersion"`
	// ID is the container ID
	ID string `json:"id"`
	// InitProcessPid is the init process id in the parent namespace
	InitProcessPid int `json:"pid"`
	// ContainerStatus is the current status of the container, running, paused, ...
	Status string `json:"status"`
	// Bundle is the path on the filesystem to the bundle
	Bundle string `json:"bundle"`
	// Rootfs is a path to a directory containing the container's root filesystem.
	Rootfs string `json:"rootfs"`
	// Created is the unix timestamp for the creation time of the container in UTC
	Created time.Time `json:"created"`
	// GetAnnotations is the user defined annotations added to the config.
	Annotations map[string]string `json:"annotations,omitempty"`
}

func GetContainerStateVOs(ids []string) ([]*ContainerStateVO, error) {
	var vos []*ContainerStateVO
	for _, id := range ids {
		vo, err := GetContainerStateVO(id)
		if err != nil {
			return nil, err
		}
		vos = append(vos, vo)
	}
	return vos, nil
}

func GetContainerStateVO(id string) (*ContainerStateVO, error) {
	container, err := GetContainer(id)
	if err != nil {
		return nil, err
	}
	state, err := container.State()
	if err != nil {
		return nil, err
	}
	containerStatus, err := container.Status()
	if err != nil {
		return nil, err
	}
	return convertContainerStateToVO(containerStatus, state), nil
}

func convertContainerStateToVO(status libcapsule.ContainerStatus, state *libcapsule.StateStorage) *ContainerStateVO {
	bundle, annotations := util.GetAnnotations(state.Config.Labels)
	return &ContainerStateVO{
		Created:        state.Created,
		Status:         status.String(),
		InitProcessPid: state.InitProcessPid,
		ID:             state.ID,
		Rootfs:         state.Config.Rootfs,
		Version:        state.Config.Version,
		Bundle:         bundle,
		Annotations:    annotations,
	}
}
