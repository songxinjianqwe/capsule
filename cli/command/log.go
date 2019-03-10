package command

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/cli/util"
	"github.com/songxinjianqwe/capsule/libcapsule/constant"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"path"
)

var LogCommand = cli.Command{
	Name:  "log",
	Usage: "get a container's log",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "exec",
			Usage: "get a container's exec log",
		},
	},
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}
		containerId := ctx.Args().First()
		_, err := util.GetContainer(containerId)
		if err != nil {
			return err
		}
		var logFilename string
		logrus.Infof("exec param: %s", ctx.String("exec"))
		if ctx.String("exec") != "" {
			// exec detach log
			logFilename = path.Join(constant.RuntimeRoot, containerId, fmt.Sprintf(constant.ContainerExecLogFilenamePattern, ctx.String("exec")))
		} else {
			// container detach log
			logFilename = path.Join(constant.RuntimeRoot, containerId, constant.ContainerInitLogFilename)
		}
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
