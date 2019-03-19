package command

import (
	"encoding/json"
	"fmt"
	"github.com/songxinjianqwe/capsule/cli/util"
	"github.com/urfave/cli"
)

var StateCommand = cli.Command{
	Name:  "state",
	Usage: "get a container's state",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "detail, d",
			Usage: "get container's detailed state",
		},
	},
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}
		vo, err := util.GetContainerStateVO(ctx.GlobalString("root"), ctx.Args().First())
		if err != nil {
			return err
		}
		if !ctx.Bool("detail") {
			vo.Detail = nil
		}
		data, err := json.MarshalIndent(vo, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}
