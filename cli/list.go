package cli

import (
	"github.com/urfave/cli"
)

var ListCommand = cli.Command{
	Name:    "list",
	Aliases: []string{"l"},
	Usage:   "list all containers",
	Action: func(c *cli.Context) error {
		return nil
	},
}
