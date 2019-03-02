package command

import (
	"github.com/songxinjianqwe/capsule/cli/util"
	"github.com/urfave/cli"
	"os"
)

var ExecCommand = cli.Command{
	Name:  "exec",
	Usage: "exec command in a container",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "detach, d",
			Usage: "detach from the container's process",
		},
	},
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}
		status, err := util.LaunchContainer(ctx.Args().First(), nil, util.ContainerActRun, false, ctx.Bool("detach"))
		if err != nil {
			return err
		}
		// 正常返回0，异常返回-1
		os.Exit(status)
		return nil
	},
}
