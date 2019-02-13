package command

import (
	"github.com/urfave/cli"
)

var CreateCommand = cli.Command{
	Name:    "create",
	Usage:   "create a containers",
	Action: func(ctx *cli.Context) error {
		return nil
	},
}
