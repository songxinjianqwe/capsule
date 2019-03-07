package command

import (
	"fmt"
	"github.com/songxinjianqwe/capsule/cli/util"
	"github.com/urfave/cli"
)

var CleanCommand = cli.Command{
	Name:  "clean",
	Usage: "clean runtime files",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "Force remove all runtime files even there are containers still exist",
		},
	},
	Action: func(ctx *cli.Context) error {
		ids, err := util.GetContainerIds()
		if err != nil {
			return err
		}
		if len(ids) > 0 && !ctx.Bool("force") {
			return fmt.Errorf("there are containers still exist, cant clean runtime files")
		}
		return util.CleanRuntime()
	},
}
