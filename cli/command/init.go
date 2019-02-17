package command

import "github.com/urfave/cli"

var InitCommand = cli.Command{
	Name:  "init",
	Usage: "init a container(init process)",
	Action: func(ctx *cli.Context) error {
		return nil
	},
}
