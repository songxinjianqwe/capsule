package command

import (
	"github.com/songxinjianqwe/rune/cli/util"
	"github.com/urfave/cli"
)

var RunCommand = cli.Command{
	Name:  "run",
	Usage: "create and start a container",
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}

		return nil
	},
}

/*
	先运行一个父进程（clone出一个namespace隔离的进程，即docker-init进程），然后重新调用自己，执行用户命令（-it）

*/
func Run(cmd string, isTty bool) {

}
