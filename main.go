package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	runeCli "github.com/songxinjianqwe/rune/cli/command"
	"github.com/urfave/cli"
	"os"
)

const (
	AppName    = "rune"
	AppVersion = "0.0.1"
	Usage      = `Open Container Initiative runtime
rune is a command line client for running applications packaged according to
the Open Container Initiative (OCI) format and is a compliant implementation of the
Open Container Initiative specification.`
)

/**
CLI入口
*/
func main() {
	app := cli.NewApp()
	app.Name = AppName
	app.Version = AppVersion
	app.Usage = Usage
	app.ExitErrHandler = func(context *cli.Context, err error) {
		logrus.Errorf(err.Error())
	}
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
	// 日志是放在文件中的，而fmt.Printf是给用户看的
	// 暂时将日志输出到stdout中
	app.Before = func(ctx *cli.Context) error {
		//设置输出样式，自带的只有两种样式logrus.JSONFormatter{}和logrus.TextFormatter{}
		formatter := new(logrus.TextFormatter)
		formatter.FullTimestamp = true                        // 显示完整时间
		formatter.TimestampFormat = "2006-01-02 15:04:05.000" // 时间格式
		formatter.DisableTimestamp = false                    // 禁止显示时间
		formatter.DisableColors = false                       // 禁止颜色显示
		logrus.SetFormatter(formatter)
		//设置output,默认为stderr,可以为任何io.Writer，比如文件*os.File
		logrus.SetOutput(os.Stdout)
		//设置最低loglevel
		logrus.SetLevel(logrus.InfoLevel)
		//设置输出文件名和行号
		logrus.SetReportCaller(true)
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
