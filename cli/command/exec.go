package command

import (
	"fmt"
	"github.com/songxinjianqwe/capsule/cli/util"
	"github.com/urfave/cli"
	"strings"
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
		if err := util.CheckArgs(ctx, 2, util.MinArgs); err != nil {
			return err
		}
		spec, err := loadSpec()
		if err != nil {
			return err
		}
		args := ctx.Args()[1:]
		if len(args) == 1 && strings.Contains(args[0], " ") {
			args = strings.Split(args[0], " ")
		}
		execId, err := util.ExecContainer(
			ctx.Args().First(),
			spec,
			ctx.Bool("detach"),
			args,
			ctx.String("cwd"),
			ctx.StringSlice("env"))
		if err != nil {
			return err
		}
		if ctx.Bool("detach") {
			fmt.Printf("exec id is %s\n", execId)
		}
		return nil
	},
}
