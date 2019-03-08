package command

import (
	"github.com/songxinjianqwe/capsule/cli/util"
	"github.com/urfave/cli"
)

/*
如果detach=true，则表示将容器的stdio、stdout、stderr设置为os.stdio...等，且等待容器进程结束
如果detach=false，则什么都不做。
并且capsule start时，detach总是为false。
*/
var RunCommand = cli.Command{
	Name:  "run",
	Usage: "create and start a container",
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
		// 将spec转为container config对象
		// 加载factory
		// 调用factory.create
		spec, err := loadSpec()
		if err != nil {
			return err
		}
		if err := util.CreateOrRunContainer(ctx.Args().First(), spec, util.ContainerActRun, ctx.Bool("detach")); err != nil {
			return err
		}
		return nil
	},
}
