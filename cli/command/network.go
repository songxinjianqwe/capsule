package command

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/cli/util"
	"github.com/songxinjianqwe/capsule/libcapsule/network"
	"github.com/urfave/cli"
	"os"
	"text/tabwriter"
)

var NetworkCommand = cli.Command{
	Name:  "network",
	Usage: "container network commands",
	Before: func(ctx *cli.Context) error {
		logrus.Infof("init network drivers...")
		if err := network.InitNetworkDrivers(ctx.GlobalString("root")); err != nil {
			return err
		}
		return nil
	},
	Subcommands: []cli.Command{
		networkCreateCommand,
		networkDeleteCommand,
		networkListCommand,
		networkShowCommand,
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
		driver := ctx.String("driver")
		if driver == "" {
			return fmt.Errorf("driver cant be empty")
		}
		subnet := ctx.String("subnet")
		if subnet == "" {
			return fmt.Errorf("subnet cant be empty")
		}
		if _, err := network.CreateNetwork(driver, subnet, ctx.Args().First()); err != nil {
			return nil
		}
		return nil
	},
}

var networkDeleteCommand = cli.Command{
	Name:  "delete",
	Usage: "delete a network",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "driver",
			Usage: "network driver",
		},
	},
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}
		driver := ctx.String("driver")
		if driver == "" {
			return fmt.Errorf("driver cant be empty")
		}
		if err := network.DeleteNetwork(driver, ctx.Args().First()); err != nil {
			return nil
		}
		return nil
	},
}

var networkListCommand = cli.Command{
	Name:  "list",
	Usage: "list networks",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "driver",
			Usage: "network driver",
		},
	},
	Action: func(ctx *cli.Context) (err error) {
		driver := ctx.String("driver")
		var networks []*network.Network
		if driver == "" {
			networks, err = network.ListAllNetwork()
		} else {
			networks, err = network.ListNetwork(driver)
		}
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
		fmt.Fprint(w, "NAME\tGATEWAY_IP\tSUBNET\tDRIVER\n")
		for _, item := range networks {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				item.Name,
				item.GatewayIP(),
				item.Subnet(),
				item.Driver,
			)
		}
		if err := w.Flush(); err != nil {
			return err
		}
		return nil
	},
}

var networkShowCommand = cli.Command{
	Name:  "show",
	Usage: "show a network",
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}
		network, err := network.LoadNetworkByName(ctx.Args().First())
		if err != nil {
			return err
		}
		fmt.Println(network)
		return nil
	},
}
