package command

import (
	"github.com/songxinjianqwe/rune/cli/util"
	"github.com/urfave/cli"
)

var StartCommand = cli.Command{
	Name:  "start",
	Usage: "start a container",
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}
		return nil
	},
}
