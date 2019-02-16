package command

import (
	"github.com/songxinjianqwe/rune/cli/util"
	"github.com/urfave/cli"
)

var CreateCommand = cli.Command{
	Name:  "create",
	Usage: "create a container",
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}
		return nil
	},
}
