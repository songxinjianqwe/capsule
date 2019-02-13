package cli

import "github.com/urfave/cli"

var CreateCommand = cli.Command{
	Name:    "create",
	Aliases: []string{"c"},
	Usage:   "create a containers",
	Action: func(c *cli.Context) error {
		return nil
	},
}
