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
		// 将spec转为container config对象
		// 加载factory
		// 调用factory.create
		spec, err := loadSpec()
		if err != nil {
			return err
		}
		ctx.Args().First()

		return nil
	},
}
