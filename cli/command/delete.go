package command

import (
	"fmt"
	"github.com/songxinjianqwe/capsule/cli/util"
	"github.com/urfave/cli"
	"golang.org/x/sys/unix"
	"syscall"
	"time"
)

var DeleteCommand = cli.Command{
	Name:  "delete",
	Usage: "delete a container",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "Forcibly deletes the container if it is still running (uses SIGKILL)",
		},
	},
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}
		container, err := util.GetContainer(ctx.Args().First())
		if err != nil {
			return err
		}
		if ctx.Bool("force") {
			_ = container.Signal(unix.SIGKILL)
			for i := 0; i < 100; i++ {
				time.Sleep(100 * time.Millisecond)
				if err := container.Signal(syscall.Signal(0)); err != nil {
					// 发信号失败说明进程已经停止
					return container.Destroy()
				}
			}
			return fmt.Errorf("waiting container dead timed out")
		} else {
			return container.Destroy()
		}
	},
}
