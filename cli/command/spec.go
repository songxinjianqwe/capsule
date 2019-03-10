package command

import (
	"encoding/json"
	"github.com/songxinjianqwe/capsule/cli/util"
	"github.com/songxinjianqwe/capsule/libcapsule/constant"
	"github.com/songxinjianqwe/capsule/libcapsule/util/spec"
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
		if err := util.CheckNoFile(constant.ContainerConfigFilename); err != nil {
			return err
		}
		data, err := json.MarshalIndent(exampleSpec, "", "\t")
		if err != nil {
			return err
		}
		return ioutil.WriteFile(constant.ContainerConfigFilename, data, 0666)
	},
}
