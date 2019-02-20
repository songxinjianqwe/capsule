package command

import (
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
			os.Exit(1)
		}
		return nil
	},
}
