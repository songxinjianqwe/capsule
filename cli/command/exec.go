package command

import "github.com/urfave/cli"

var ExecCommand = cli.Command{
	Name:  "exec",
	Usage: "exec command in a container",
	Action: func(ctx *cli.Context) error {
		return nil
	},
}
