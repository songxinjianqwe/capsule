package command

import "github.com/urfave/cli"

var KillCommand = cli.Command{
	Name:  "kill",
	Usage: "kill a container",
	Action: func(ctx *cli.Context) error {
		return nil
	},
}
