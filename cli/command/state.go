package command

import (
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/cli/util"
	"github.com/songxinjianqwe/rune/libcapsule"
	"github.com/urfave/cli"
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
	// The owner of the state directory (the owner of the container).
	Owner string `json:"owner"`
}

var StateCommand = cli.Command{
	Name:  "state",
	Usage: "get a container's state",
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}
		container, err := util.GetContainer(ctx.Args().First())
		if err != nil {
			return err
		}
		state, err := container.State()
		if err != nil {
			return err
		}
		logrus.Info(convertContainerStateToVO(state))
		return nil
	},
}

func convertContainerStateToVO(state *libcapsule.StateStorage) ContainerStateVO {
	panic("implement me")
}
