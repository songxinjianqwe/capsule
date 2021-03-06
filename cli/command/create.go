package command

import (
	"github.com/songxinjianqwe/capsule/cli/util"
	"github.com/songxinjianqwe/capsule/libcapsule/facade"
	"github.com/urfave/cli"
)

var CreateCommand = cli.Command{
	Name:  "create",
	Usage: "create a container",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "bundle, b",
			Value: "",
			Usage: `path to the root of the bundle directory, defaults to the current directory`,
		},
		cli.StringFlag{
			Name:  "network, net",
			Usage: `network connected by container`,
		},
		cli.StringSliceFlag{
			Name:  "port, p",
			Usage: `port mappings, example: host port:container port`,
		},
	},
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}
		// 将spec转为container config对象
		// 加载factory
		// 调用factory.create
		spec, err := facade.LoadSpec(ctx.String("bundle"))
		if err != nil {
			return err
		}
		if err := facade.CreateOrRunContainer(ctx.GlobalString("root"), ctx.Args().First(), ctx.String("bundle"), spec, facade.ContainerActCreate, false, ctx.String("network"), ctx.StringSlice("port")); err != nil {
			return err
		}
		return nil
	},
}
