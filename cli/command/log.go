package command

import (
	"fmt"
	"github.com/songxinjianqwe/capsule/cli/util"
	"github.com/songxinjianqwe/capsule/libcapsule"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"path"
)

var LogCommand = cli.Command{
	Name:  "log",
	Usage: "get a container's log",
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}
		containerId := ctx.Args().First()
		_, err := util.GetContainer(containerId)
		if err != nil {
			return err
		}
		logFilename := path.Join(libcapsule.RuntimeRoot, containerId, libcapsule.ContainerLogFilename)
		file, err := os.Open(logFilename)
		if err != nil {
			return err
		}
		bytes, err := ioutil.ReadAll(file)
		if err != nil {
			return err
		}
		fmt.Print(string(bytes))
		return nil
	},
}
