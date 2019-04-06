package command

import (
	"encoding/json"
	"fmt"
	"github.com/songxinjianqwe/capsule/cli/util"
	"github.com/songxinjianqwe/capsule/libcapsule/facade"
	"github.com/songxinjianqwe/capsule/libcapsule/image"
	"github.com/songxinjianqwe/capsule/libcapsule/network"
	"github.com/urfave/cli"
	"golang.org/x/sys/unix"
	"os"
	"strings"
	"text/tabwriter"
	"time"
)

var ImageCommand = cli.Command{
	Name:  "image",
	Usage: "container image commands",
	Subcommands: []cli.Command{
		imageCreateCommand,
		imageDeleteCommand,
		imageListCommand,
		imageGetCommand,
		imageRunContainerCommand,
		imageDestroyContainerCommand,
	},
}

var imageCreateCommand = cli.Command{
	Name:  "create",
	Usage: "create a image, like create $image_name $tarPath",
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 2, util.ExactArgs); err != nil {
			return err
		}
		// 这里是将一个存在的tar文件作为新的rootfs
		imageService, err := image.NewImageService(ctx.GlobalString("root"))
		if err != nil {
			return err
		}
		if err := imageService.Create(ctx.Args().First(), ctx.Args()[1]); err != nil {
			return err
		}
		return nil
	},
}

var imageDeleteCommand = cli.Command{
	Name:  "delete",
	Usage: "delete a image",
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}
		imageService, err := image.NewImageService(ctx.GlobalString("root"))
		if err != nil {
			return err
		}
		if err := imageService.Delete(ctx.Args().First()); err != nil {
			return err
		}
		return nil
	},
}

var imageListCommand = cli.Command{
	Name:  "list",
	Usage: "list images",
	Action: func(ctx *cli.Context) error {
		imageService, err := image.NewImageService(ctx.GlobalString("root"))
		if err != nil {
			return err
		}
		images, err := imageService.List()
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
		fmt.Fprint(w, "ID\tLAYER_ID\tCREATED\tSIZE\n")
		for _, item := range images {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				item.Id,
				item.LayerId,
				item.CreateTime.Format(time.RFC3339Nano),
				fmt.Sprintf("%dMB", item.Size))
		}
		if err := w.Flush(); err != nil {
			return err
		}
		return nil
	},
}

var imageGetCommand = cli.Command{
	Name:  "get",
	Usage: "get a image",
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}
		imageService, err := image.NewImageService(ctx.GlobalString("root"))
		if err != nil {
			return err
		}
		image, err := imageService.Get(ctx.Args().First())
		if err != nil {
			return err
		}
		data, err := json.MarshalIndent(image, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	},
}

// image run $image_name command
// -d
// -workdir $workdir
// -hostname $hostname
// -name $name
// -env a=b c=d
// -cpushare
// -memory
// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~下面是spec里没有的,由capsule负责做的配置信息
// -link
// -volume a/a:b
// -network $network_name
// -port xx:xxx
// -label a=b
var imageRunContainerCommand = cli.Command{
	Name:  "run",
	Usage: "run container in image way",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "detach, d",
			Usage: "detach from the container's process",
		},
		cli.StringFlag{
			Name:  "id",
			Usage: "container unique id",
		},
		cli.StringFlag{
			Name:  "cwd",
			Value: "/",
			Usage: "current work directory",
		},
		cli.StringSliceFlag{
			Name:  "env, e",
			Usage: "environment variables",
		},
		cli.StringFlag{
			Name:  "hostname",
			Usage: "hostname",
		},
		cli.Int64Flag{
			Name:  "cpushare",
			Value: 1024,
			Usage: "cpushare",
		},
		cli.Int64Flag{
			Name:  "memory",
			Value: 0,
			Usage: "memory limit",
		},
		cli.StringFlag{
			Name:  "network",
			Value: network.DefaultBridgeName,
			Usage: "network name",
		},
		cli.StringSliceFlag{
			Name:  "port, p",
			Usage: "port mappings",
		},
		cli.StringSliceFlag{
			Name:  "label, l",
			Usage: "container label",
		},
	},
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 2, util.MinArgs); err != nil {
			return err
		}
		imageService, err := image.NewImageService(ctx.GlobalString("root"))
		if err != nil {
			return err
		}
		containerId := ctx.String("id")
		if containerId == "" {
			return fmt.Errorf("container id can not be empty")
		}
		hostname := ctx.String("hostname")
		if hostname == "" {
			hostname = containerId
		}
		annotations := make(map[string]string)
		for _, label := range ctx.StringSlice("label") {
			splits := strings.SplitN(label, "=", 2)
			annotations[splits[0]] = splits[1]
		}
		if err := imageService.Run(&image.ImageRunArgs{
			ImageId:      ctx.Args().First(),
			ContainerId:  containerId,
			Args:         ctx.Args()[1:],
			Env:          ctx.StringSlice("env"),
			Cwd:          ctx.String("cwd"),
			Hostname:     hostname,
			Cpushare:     ctx.Uint64("cpushare"),
			Memory:       ctx.Int64("memory"),
			Annotations:  annotations,
			Network:      ctx.String("network"),
			PortMappings: ctx.StringSlice("port"),
			Detach:       ctx.Bool("detach"),
		}); err != nil {
			return err
		}
		return nil
	},
}

var imageDestroyContainerCommand = cli.Command{
	Name:  "destroy",
	Usage: "destroy container in image way",
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "force, f",
			Usage: "Force delete the container even it is still running (uses SIGKILL)",
		},
	},
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.ExactArgs); err != nil {
			return err
		}
		imageService, err := image.NewImageService(ctx.GlobalString("root"))
		if err != nil {
			return err
		}
		container, err := facade.GetContainer(ctx.GlobalString("root"), ctx.Args().First())
		if err != nil {
			return err
		}
		if ctx.Bool("force") {
			_ = container.Signal(unix.SIGKILL)
			for i := 0; i < 100; i++ {
				time.Sleep(100 * time.Millisecond)
				return imageService.Destroy(container)
			}
			return fmt.Errorf("waiting container dead timed out")
		} else {
			return imageService.Destroy(container)
		}
	},
}
