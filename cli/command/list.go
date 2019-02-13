package command

import (
	"github.com/urfave/cli"
)

var ListCommand = cli.Command{
	Name:    "list",
	Usage:   "list all containers",
	Action: func(c *cli.Context) error {
		return nil
	},
}
