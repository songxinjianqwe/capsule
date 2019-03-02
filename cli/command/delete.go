package command

import (
	"github.com/songxinjianqwe/capsule/cli/util"
	"github.com/urfave/cli"
)

var DeleteCommand = cli.Command{
	Name:  "delete",
	Usage: "delete a container",
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}
		container, err := util.GetContainer(ctx.Args().First())
		if err != nil {
			return err
		}
		return container.Destroy()
	},
}
