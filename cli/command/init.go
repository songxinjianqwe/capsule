package command

import (
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule"
	_ "github.com/songxinjianqwe/capsule/libcapsule/nsenter"
	"github.com/urfave/cli"
	"os"
	"runtime"
)

func init() {
	if len(os.Args) > 1 && os.Args[1] == "init" {
		logrus.Infof("setting go max procs = 1")
		runtime.GOMAXPROCS(1)
		runtime.LockOSThread()
	}
}

/*
仅限内部调用，有可能是init process，也有可能是exec process。
*/
var InitCommand = cli.Command{
	Name:  "init",
	Usage: "init a container(execute init/exec process)",
	Action: func(ctx *cli.Context) error {
		factory, err := libcapsule.NewFactory(ctx.GlobalString("root"), false)
		if err != nil {
			return err
		}
		if err := factory.StartInitialization(); err != nil {
			logrus.WithField("init", true).Errorf("init failed, err: %s", err.Error())
		}
		return nil
	},
}
