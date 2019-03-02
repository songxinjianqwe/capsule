package command

import (
	"github.com/songxinjianqwe/capsule/cli/util"
	"github.com/urfave/cli"
)

var LogCommand = cli.Command{
	Name:  "log",
	Usage: "get a container's log",
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}
		return nil
	},
}
