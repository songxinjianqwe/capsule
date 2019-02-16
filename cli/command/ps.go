package command

import "github.com/urfave/cli"

var PsCommand = cli.Command{
	Name:  "ps",
	Usage: "show a container's info",
	Action: func(ctx *cli.Context) error {
		return nil
	},
}
