package command

import "github.com/urfave/cli"

var DeleteCommand = cli.Command{
	Name:  "delete",
	Usage: "delete a container",
	Action: func(ctx *cli.Context) error {
		return nil
	},
}
