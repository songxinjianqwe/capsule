package command

import (
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule"
	"github.com/urfave/cli"
	"os"
)

var InitCommand = cli.Command{
	Name:  "init",
	Usage: "init a container(execute init process)",
	Action: func(ctx *cli.Context) error {
		factory, _ := libcapsule.NewFactory()
		if err := factory.StartInitialization(); err != nil {
			logrus.WithField("init", true).Errorf("init failed, err: %s", err.Error())
			os.Exit(1)
		}
		return nil
	},
}
