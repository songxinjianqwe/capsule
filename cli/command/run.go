package command

import (
	"github.com/songxinjianqwe/rune/cli/util"
	"github.com/urfave/cli"
	"os"
)

/**
如果Terminal=true，则表示将容器的stdio、stdout、stderr设置为os.stdio...等，且等待容器进程结束
如果Terminal=false，则什么都不做。
*/
var RunCommand = cli.Command{
	Name:  "run",
	Usage: "create and start a container",
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
		status, err := util.LaunchContainer(ctx.Args().First(), spec, util.ContainerActRun)
		if err != nil {
			return err
		}
		// 正常返回0，异常返回-1
		os.Exit(status)
		return nil
	},
}
