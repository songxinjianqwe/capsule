package libcapsule

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/configs"
	"github.com/songxinjianqwe/capsule/libcapsule/util"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
	"github.com/songxinjianqwe/capsule/libcapsule/util/rootfs"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
)

type InitializerStandardImpl struct {
	config     *InitExecConfig
	configPipe *os.File
	parentPid  int
}

// **************************************************************************************************
// public
// **************************************************************************************************

/*
容器初始化
*/
func (initializer *InitializerStandardImpl) Init() (err error) {
	logrus.WithField("init", true).Infof("InitializerStandardImpl Init()")
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

	// 初始化rootfs
	if err = initializer.setUpRootfs(); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.RootfsError, "init process/prepare rootfs")
	}

	// 如果有设置Mount的Namespace，则设置rootfs与mount为read only（如果需要的话）
	if initializer.config.ContainerConfig.Namespaces.Contains(configs.NEWNS) {
		if err := initializer.SetRootfsReadOnlyIfSpecified(); err != nil {
			return err
		}
	}

	// 初始化hostname
	if hostname := initializer.config.ContainerConfig.Hostname; hostname != "" {
		logrus.WithField("init", true).Infof("init process/setting hostname: %s", hostname)
		if err = unix.Sethostname([]byte(hostname)); err != nil {
			return exception.NewGenericErrorWithContext(err, exception.HostnameError, "init process/set hostname")
		}
	}

	// 初始化环境变量
	for key, value := range initializer.config.ContainerConfig.Sysctl {
		if err = writeSystemProperty(key, value); err != nil {
			return exception.NewGenericErrorWithContext(err, exception.SysctlError, fmt.Sprintf("init process/write sysctl key %s", key))
		}
	}

	// look path 可以在系统的PATH里面寻找命令的绝对路径
	name, err := exec.LookPath(initializer.config.ProcessConfig.Args[0])
	if err != nil {
		return exception.NewGenericErrorWithContext(err, exception.LookPathError, "init process/look path cmd")
	}
	logrus.WithField("init", true).Infof("look path: %s", name)

	logrus.WithField("init", true).Infof("sync parent ready...")
	// child --------------> parent
	// 告诉parent，init process已经初始化完毕，马上要执行命令了
	if err := util.SyncSignal(initializer.parentPid, syscall.SIGUSR1); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SignalError, "init process/sync parent ready")
	}

	// child <-------------- parent
	// 等待parent给一个继续执行命令，即exec的信号
	logrus.WithField("init", true).Info("start waiting parent continue(SIGUSR2) signal...")
	util.WaitSignal(syscall.SIGUSR2)
	logrus.WithField("init", true).Info("received parent continue(SIGUSR2) signal")

	logrus.WithField("init", true).Info("execute real command and cover capsule init config")
	// syscall.Exec与cmd.Start不同，后者是启动一个新的进程来执行命令
	// 而前者会在覆盖当前进程的镜像、数据、堆栈等信息，包括PID。
	logrus.WithField("init", true).Infof("syscall.Exec(name: %s, args: %v, env: %v)...", name, initializer.config.ProcessConfig.Args, os.Environ())
	if err := syscall.Exec(name, initializer.config.ProcessConfig.Args, os.Environ()); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.SyscallExecuteCmdError, "start init process")
	}
	return nil
}

// **************************************************************************************************
// private
// **************************************************************************************************

func (initializer *InitializerStandardImpl) setUpRootfs() error {
	logrus.WithField("init", true).Info("setting up rootfs...")
	containerRootfs := initializer.config.ContainerConfig.Rootfs
	if err := rootfs.PrepareRoot(&initializer.config.ContainerConfig); err != nil {
		return exception.NewGenericErrorWithContext(err, exception.PrepareRootError, "preparing root")
	}

	// 挂载
	for _, m := range initializer.config.ContainerConfig.Mounts {
		if err := rootfs.MountToRootfs(m, containerRootfs); err != nil {
			return exception.NewGenericErrorWithContext(err, exception.MountError, fmt.Sprintf("mounting %q to rootfs %q at %q", m.Source, initializer.config.ContainerConfig.Rootfs, m.Destination))
		}
	}
	// 设备
	for _, node := range initializer.config.ContainerConfig.Devices {
		// containers running in a user namespace are not allowed to mknod
		// devices so we can just bind mount it from the host.
		if err := rootfs.CreateDeviceNode(containerRootfs, node); err != nil {
			return err
		}
	}
	// 如果使用了Mount的namespace，则使用pivot_root命令
	// pivot root放在mount之前的话，会报错invalid argument
	if initializer.config.ContainerConfig.Namespaces.Contains(configs.NEWNS) {
		if err := rootfs.PivotRoot(containerRootfs); err != nil {
			return err
		}
	}
	return nil
}

func (initializer *InitializerStandardImpl) SetRootfsReadOnlyIfSpecified() error {
	// remount to set read only if specified
	for _, m := range initializer.config.ContainerConfig.Mounts {
		// 仅针对/dev,这个目录应该是只读
		if util.CleanPath(m.Destination) == "/dev" {
			if m.Flags&unix.MS_RDONLY == unix.MS_RDONLY {
				if err := rootfs.RemountReadonly(m); err != nil {
					return exception.NewGenericErrorWithContext(err, exception.MountError, fmt.Sprintf("remounting %q as readonly", m.Destination))
				}
			}
			break
		}
	}

	// set rootfs ( / ) as readonly
	if initializer.config.ContainerConfig.Readonlyfs {
		if err := rootfs.SetRootfsReadonly(); err != nil {
			return exception.NewGenericErrorWithContext(err, exception.MountError, "setting rootfs as readonly")
		}
	}
	return nil
}

// **************************************************************************************************
// util
// **************************************************************************************************

// writeSystemProperty writes the value to a path under /proc/sys as determined from the key.
// For e.g. net.ipv4.ip_forward translated to /proc/sys/net/ipv4/ip_forward.
func writeSystemProperty(key string, value string) error {
	keyPath := strings.Replace(key, ".", "/", -1)
	logrus.WithField("init", true).Infof("write system property: key:%s, value:%s", keyPath, value)
	return ioutil.WriteFile(path.Join("/proc/sys", keyPath), []byte(value), 0644)
}
