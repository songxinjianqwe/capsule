package command

import (
	"github.com/songxinjianqwe/capsule/cli/util"
	"github.com/urfave/cli"
)

/*
相当于exec(ps -ef)
*/
var PsCommand = cli.Command{
	Name:  "ps",
	Usage: "show a container's process info",
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}
		if _, err := util.ExecContainer(
			ctx.Args().First(),
			false,
			[]string{"ps", "-ef"},
			"",
			nil); err != nil {
			return err
		}
		return nil
	},
}
