package command

import (
	"fmt"
	"github.com/songxinjianqwe/capsule/libcapsule/facade"
	"github.com/urfave/cli"
	"os"
	"text/tabwriter"
	"time"
)

var ListCommand = cli.Command{
	Name:  "list",
	Usage: "list all containers",
	Action: func(ctx *cli.Context) error {
		ids, err := facade.GetContainerIds(ctx.GlobalString("root"))
		if err != nil {
			return err
		}
		vos, err := facade.GetContainerStateVOs(ctx.GlobalString("root"), ids)
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
		fmt.Fprint(w, "ID\tPID\tSTATUS\tIP\tBUNDLE\tCREATED\n")
		for _, item := range vos {
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\t%s\n",
				item.ID,
				item.InitProcessPid,
				item.Status,
				item.IP,
				item.Bundle,
				item.Created.Format(time.RFC3339Nano))
		}
		if err := w.Flush(); err != nil {
			return err
		}
		return nil
	},
}
