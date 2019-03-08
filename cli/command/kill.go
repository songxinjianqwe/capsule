package command

import (
	"fmt"
	"github.com/songxinjianqwe/capsule/cli/util"
	"github.com/urfave/cli"
	"golang.org/x/sys/unix"
	"strconv"
	"strings"
	"syscall"
)

/*
如果杀死容器的init process，那么容器的exec process也会被杀死
暂不清楚原因。
*/
var KillCommand = cli.Command{
	Name:  "kill",
	Usage: "kill a container",
	Action: func(ctx *cli.Context) error {
		if err := util.CheckArgs(ctx, 1, util.MinArgs); err != nil {
			return err
		}
		if err := util.CheckArgs(ctx, 2, util.MaxArgs); err != nil {
			return err
		}
		container, err := util.GetContainer(ctx.Args().First())
		if err != nil {
			return err
		}

		sigStr := ctx.Args().Get(1)
		if sigStr == "" {
			sigStr = "SIGTERM"
		}
		signal, err := parseSignal(sigStr)
		if err != nil {
			return err
		}
		return container.Signal(signal)
	},
}

func parseSignal(rawSignal string) (syscall.Signal, error) {
	s, err := strconv.Atoi(rawSignal)
	if err == nil {
		return syscall.Signal(s), nil
	}
	signal, ok := signalMap[strings.TrimPrefix(strings.ToUpper(rawSignal), "SIG")]
	if !ok {
		return -1, fmt.Errorf("unknown signal %q", rawSignal)
	}
	return signal, nil
}

var signalMap = map[string]syscall.Signal{
	"ABRT":   unix.SIGABRT,
	"ALRM":   unix.SIGALRM,
	"BUS":    unix.SIGBUS,
	"CHLD":   unix.SIGCHLD,
	"CLD":    unix.SIGCLD,
	"CONT":   unix.SIGCONT,
	"FPE":    unix.SIGFPE,
	"HUP":    unix.SIGHUP,
	"ILL":    unix.SIGILL,
	"INT":    unix.SIGINT,
	"IO":     unix.SIGIO,
	"IOT":    unix.SIGIOT,
	"KILL":   unix.SIGKILL,
	"PIPE":   unix.SIGPIPE,
	"POLL":   unix.SIGPOLL,
	"PROF":   unix.SIGPROF,
	"PWR":    unix.SIGPWR,
	"QUIT":   unix.SIGQUIT,
	"SEGV":   unix.SIGSEGV,
	"STKFLT": unix.SIGSTKFLT,
	"STOP":   unix.SIGSTOP,
	"SYS":    unix.SIGSYS,
	"TERM":   unix.SIGTERM,
	"TRAP":   unix.SIGTRAP,
	"TSTP":   unix.SIGTSTP,
	"TTIN":   unix.SIGTTIN,
	"TTOU":   unix.SIGTTOU,
	"URG":    unix.SIGURG,
	"USR1":   unix.SIGUSR1,
	"USR2":   unix.SIGUSR2,
	"VTALRM": unix.SIGVTALRM,
	"WINCH":  unix.SIGWINCH,
	"XCPU":   unix.SIGXCPU,
	"XFSZ":   unix.SIGXFSZ,
}
