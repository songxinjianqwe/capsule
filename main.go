package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	cliCmd "github.com/songxinjianqwe/scheduler/cli/command"
	daemonCmd "github.com/songxinjianqwe/scheduler/daemon/command"
	"github.com/urfave/cli"
	"os"
)

const (
	appName    = "scheduler"
	appVersion = "1.0.0"
)

/**
主goroutine从命令行读入类型、延迟时间与打印内容，并构造延迟任务或定时任务并使用timer来调度
*/
func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Version = appVersion
	app.Commands = []cli.Command{
		cliCmd.GetCommand,
		cliCmd.ListCommand,
		cliCmd.SubmitCommand,
		cliCmd.StopCommand,
		cliCmd.DeleteCommand,
		daemonCmd.DaemonCommand,
	}
	app.Before = func(c *cli.Context) error {
		//设置输出样式，自带的只有两种样式logrus.JSONFormatter{}和logrus.TextFormatter{}
		log.SetFormatter(&log.TextFormatter{})
		//设置output,默认为stderr,可以为任何io.Writer，比如文件*os.File
		log.SetOutput(os.Stdout)
		//设置最低loglevel
		log.SetLevel(log.InfoLevel)
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
