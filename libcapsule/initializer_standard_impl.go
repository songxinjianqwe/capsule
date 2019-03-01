package libcapsule

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"github.com/songxinjianqwe/rune/libcapsule/util/rootfs"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

type InitializerStandardImpl struct {
	config     *InitConfig
	configPipe *os.File
	parentPid  int
}

// **************************************************************************************************
// public
// **************************************************************************************************

/**
容器初始化
*/
func (initializer *InitializerStandardImpl) Init() (err error) {
	util.PrintSubsystemPids("memory", initializer.config.ID, "before initializer init", true)
	logrus.WithField("init", true).Infof("InitializerStandardImpl create to Init()")
	defer func() {
		// 后面再出现err就不管了
		if err != nil {
			logrus.WithField("init", true).Errorf("init process failed, send SIGCHLD signal to parent")
			// 告诉parent，init process已经初始化完毕，马上要执行命令了
			if err := unix.Kill(initializer.parentPid, syscall.SIGCHLD); err != nil {
				logrus.WithField("init", true).Errorf("send SIGABRT signal to parent failed")
			}
		}
	}()
	// 初始化网络
	if err = initializer.setUpNetwork(); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "init config/set up network")
	}

	// 初始化路由
	if err = initializer.setUpRoute(); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "init config/set up route")
	}

	// 初始化rootfs
	if err = initializer.setUpRootfs(); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "init config/prepare rootfs")
	}

	// 如果有设置Mount的Namespace，则设置rootfs与mount为read only（如果需要的话）
	if initializer.config.ContainerConfig.Namespaces.Contains(configc.NEWNS) {
		if err := initializer.SetRootfsReadOnlyIfNeed(); err != nil {
			return err
		}
	}
	util.PrintSubsystemPids("memory", initializer.config.ID, "after rootfs set up", true)

	// 初始化hostname
	if hostname := initializer.config.ContainerConfig.Hostname; hostname != "" {
		logrus.WithField("init", true).Infof("setting hostname: %s", hostname)
		if err = unix.Sethostname([]byte(hostname)); err != nil {
			return util.NewGenericErrorWithContext(err, util.SystemError, "init config/set hostname")
		}
	}

	// 初始化环境变量
	for key, value := range initializer.config.ContainerConfig.Sysctl {
		if err = writeSystemProperty(key, value); err != nil {
			return util.NewGenericErrorWithContext(err, util.SystemError, fmt.Sprintf("write sysctl key %s", key))
		}
	}

	// 初始化namespace
	if err = initializer.finalizeNamespace(); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "init config/finalize namespace")
	}

	// look path 可以在系统的PATH里面寻找命令的绝对路径
	name, err := exec.LookPath(initializer.config.ProcessConfig.Args[0])
	if err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "init config/look path cmd")
	}
	logrus.WithField("init", true).Infof("look path: %s", name)

	logrus.WithField("init", true).Infof("sync parent ready...")
	// child --------------> parent
	// 告诉parent，init process已经初始化完毕，马上要执行命令了
	if err := unix.Kill(initializer.parentPid, syscall.SIGUSR1); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "init config/sync parent ready")
	}
	// child <-------------- parent
	// 等待parent给一个继续执行命令，即exec的信号
	logrus.WithField("init", true).Info("start waiting parent continue(SIGUSR2) signal...")
	receivedChan := make(chan os.Signal, 1)
	signal.Notify(receivedChan, syscall.SIGUSR2)
	<-receivedChan
	logrus.WithField("init", true).Info("received parent continue(SIGUSR2) signal")

	logrus.WithField("init", true).Info("execute real command and cover rune init config")
	// syscall.Exec与cmd.Start不同，后者是启动一个新的进程来执行命令
	// 而前者会在覆盖当前进程的镜像、数据、堆栈等信息，包括PID。
	logrus.WithField("init", true).Infof("syscall.Exec(name: %s, args: %v, env: %v)...", name, initializer.config.ProcessConfig.Args, os.Environ())

	if err := syscall.Exec(name, initializer.config.ProcessConfig.Args, os.Environ()); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "start user config")
	}
	return nil
}

// **************************************************************************************************
// private
// **************************************************************************************************

func (initializer *InitializerStandardImpl) setUpNetwork() error {
	logrus.WithField("init", true).Info("setting up network...")
	return nil
}

func (initializer *InitializerStandardImpl) setUpRoute() error {
	logrus.WithField("init", true).Info("setting up route...")
	return nil
}

func (initializer *InitializerStandardImpl) setUpRootfs() error {
	logrus.WithField("init", true).Info("setting up rootfs...")

	if err := rootfs.PrepareRoot(&initializer.config.ContainerConfig); err != nil {
		return util.NewGenericErrorWithContext(err, util.SystemError, "preparing root")
	}
	// 挂载
	for _, m := range initializer.config.ContainerConfig.Mounts {
		if err := rootfs.MountToRootfs(m, initializer.config.ContainerConfig.Rootfs); err != nil {
			return util.NewGenericErrorWithContext(err, util.SystemError, fmt.Sprintf("mounting %q to rootfs %q at %q", m.Source, initializer.config.ContainerConfig.Rootfs, m.Destination))
		}
	}

	// pivot root放在mount之前的话，会报错invalid argument
	// 如果使用了Mount的namespace，则使用pivot_root命令
	if initializer.config.ContainerConfig.Namespaces.Contains(configc.NEWNS) {
		if err := rootfs.PivotRoot(initializer.config.ContainerConfig.Rootfs); err != nil {
			return err
		}
	}
	return nil
}

func (initializer *InitializerStandardImpl) finalizeNamespace() error {
	logrus.WithField("init", true).Info("finalizing namespace...")
	return nil
}

func (initializer *InitializerStandardImpl) SetRootfsReadOnlyIfNeed() error {
	// remount dev as ro if specified
	for _, m := range initializer.config.ContainerConfig.Mounts {
		if util.CleanPath(m.Destination) == "/dev" {
			if m.Flags&unix.MS_RDONLY == unix.MS_RDONLY {
				if err := rootfs.RemountReadonly(m); err != nil {
					return util.NewGenericErrorWithContext(err, util.SystemError, fmt.Sprintf("remounting %q as readonly", m.Destination))
				}
			}
			break
		}
	}

	// set rootfs ( / ) as readonly
	if initializer.config.ContainerConfig.Readonlyfs {
		if err := rootfs.SetRootfsReadonly(); err != nil {
			return util.NewGenericErrorWithContext(err, util.SystemError, "setting rootfs as readonly")
		}
	}
	return nil
}

// **************************************************************************************************
// util
// **************************************************************************************************

func writeSystemProperty(key string, value string) error {
	logrus.WithField("init", true).Infof("write system property:key:%s, value:%s", key, value)
	return nil
}
