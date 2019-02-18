package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	runeCli "github.com/songxinjianqwe/rune/cli/command"
	"github.com/songxinjianqwe/rune/cli/constant"
	"github.com/urfave/cli"
	"os"
)

/**
CLI入口
*/
func main() {
	app := cli.NewApp()
	app.Name = constant.AppName
	app.Version = constant.AppVersion
	app.Usage = constant.Usage
	app.Commands = []cli.Command{
		runeCli.CreateCommand,
		runeCli.StartCommand,
		runeCli.RunCommand,
		runeCli.ListCommand,
		runeCli.DeleteCommand,
		runeCli.ExecCommand,
		runeCli.InitCommand,
		runeCli.KillCommand,
		runeCli.PsCommand,
		runeCli.StateCommand,
		runeCli.SpecCommand,
	}
	app.Before = func(c *cli.Context) error {
		//设置输出样式，自带的只有两种样式logrus.JSONFormatter{}和logrus.TextFormatter{}
		logrus.SetFormatter(&logrus.TextFormatter{})
		//设置output,默认为stderr,可以为任何io.Writer，比如文件*os.File
		logrus.SetOutput(os.Stdout)
		//设置最低loglevel
		logrus.SetLevel(logrus.InfoLevel)
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
