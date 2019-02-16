package command

import (
	"encoding/json"
	"github.com/songxinjianqwe/rune/cli/constant"
	"github.com/songxinjianqwe/rune/cli/util"
	"github.com/songxinjianqwe/rune/libcapsule/spec"
	"github.com/urfave/cli"
	"io/ioutil"
)

var SpecCommand = cli.Command{
	Name:  "spec",
	Usage: "create a new specification file",
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 0, util.ExactArgs); err != nil {
			return err
		}
		exampleSpec := spec.Example()
		if err := util.CheckNoFile(constant.SpecConfig); err != nil {
			return err
		}
		data, err := json.MarshalIndent(exampleSpec, "", "\t")
		if err != nil {
			return err
		}
		return ioutil.WriteFile(constant.SpecConfig, data, 0666)
	},
}
