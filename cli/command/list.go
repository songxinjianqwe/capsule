package command

import (
	"fmt"
	"github.com/songxinjianqwe/rune/cli/util"
	"github.com/urfave/cli"
	"os"
	"text/tabwriter"
	"time"
)

var ListCommand = cli.Command{
	Name:  "list",
	Usage: "list all containers",
	Action: func(c *cli.Context) error {
		ids, err := util.GetContainerIds()
		if err != nil {
			return err
		}
		vos, err := util.GetContainerStateVOs(ids)
		if err != nil {
			return err
		}
		w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
		fmt.Fprint(w, "ID\tPID\tSTATUS\tBUNDLE\tCREATED\n")
		for _, item := range vos {
			fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\n",
				item.ID,
				item.InitProcessPid,
				item.Status,
				item.Bundle,
				item.Created.Format(time.RFC3339Nano))
		}
		if err := w.Flush(); err != nil {
			return err
		}
		return nil
	},
}
