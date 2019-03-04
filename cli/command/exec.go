package command

import (
	"fmt"
	"github.com/songxinjianqwe/capsule/cli/util"
	"github.com/urfave/cli"
)

var ExecCommand = cli.Command{
	Name:  "exec",
	Usage: "exec command in a container",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "detach, d",
			Usage: "detach from the container's process",
		},
		cli.StringSliceFlag{
			Name:  "env, e",
			Usage: "set environment variables",
		},
		cli.StringFlag{
			Name:  "cwd",
			Usage: "current work directory of exec process",
		},
	},
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.MinArgs); err != nil {
			return err
		}
		if len(ctx.Args()) == 1 {
			return fmt.Errorf("process args cannot be empty")
		}
		spec, err := loadSpec()
		if err != nil {
			return err
		}
		if err := util.ExecContainer(
			ctx.Args().First(),
			spec,
			ctx.Bool("detach"),
			ctx.Args()[1:],
			ctx.String("cwd"),
			ctx.StringSlice("env")); err != nil {
			return err
		}
		return nil
	},
}
