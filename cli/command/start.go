package command

import (
	"errors"
	"fmt"
	"github.com/songxinjianqwe/rune/cli/util"
	"github.com/songxinjianqwe/rune/libcapsule"
	"github.com/urfave/cli"
)

var StartCommand = cli.Command{
	Name:  "start",
	Usage: "start a container",
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}
		container, err := util.GetContainer(ctx.Args().First())
		if err != nil {
			return err
		}
		status, err := container.Status()
		if err != nil {
			return err
		}
		switch status {
		case libcapsule.Created:
			return container.Start()
		case libcapsule.Stopped:
			return errors.New("cannot start a container that has stopped")
		case libcapsule.Running:
			return errors.New("cannot start an already running container")
		default:
			return fmt.Errorf("cannot start a container in the %s state\n", status)
		}
	},
}
