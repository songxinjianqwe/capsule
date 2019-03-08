package command

import "github.com/urfave/cli"

/*
相当于exec ps
*/
var PsCommand = cli.Command{
	Name:  "ps",
	Usage: "show a container's process info",
	Action: func(ctx *cli.Context) error {

		return nil
	},
}
