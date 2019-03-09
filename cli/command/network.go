package command

import (
	"github.com/songxinjianqwe/capsule/cli/util"
	"github.com/urfave/cli"
)

var NetworkCommand = cli.Command{
	Name:  "network",
	Usage: "container network commands",
	Subcommands: []cli.Command{
		networkCreateCommand,
		networkDeleteCommand,
		networkListCommand,
	},
}

var networkCreateCommand = cli.Command{
	Name:  "create",
	Usage: "create a network",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "driver",
			Usage: "network driver",
		},
		cli.StringFlag{
			Name:  "subnet",
			Usage: "subnet cidr",
		},
	},
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}

		return nil
	},
}

var networkDeleteCommand = cli.Command{
	Name:  "delete",
	Usage: "delete a network",
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}
		return nil
	},
}

var networkListCommand = cli.Command{
	Name:  "list",
	Usage: "list networks",
	Action: func(ctx *cli.Context) error {
		return nil
	},
}
