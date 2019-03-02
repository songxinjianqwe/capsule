package rootfs

import (
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule/configc"
	"github.com/songxinjianqwe/capsule/libcapsule/util"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"strings"
)

func PrepareRoot(config *configc.ContainerConfig) error {
	logrus.WithField("init", true).Info("preparing root...")
	flag := unix.MS_SLAVE | unix.MS_REC
	logrus.WithField("init", true).Info("mounting / in \"\" fs...")
	// 这行必须要有，否则在某些情况下宿主机会出现 open /dev/null: no such file or directory
	if err := unix.Mount("", "/", "", uintptr(flag), ""); err != nil {
		return err
	}
	logrus.WithField("init", true).Infof("mounting %s in bind fs...", config.Rootfs)
	// https://unix.stackexchange.com/questions/424478/bind-mounting-source-to-itself
	// 自己bind自己是为了后面的设置成read only，可以加一些mount的flag
	// 另一方面，还可以阻止文件被移动或者被链接

	// 可以让当前root的老root和新root不在同一个文件系统
	// bind mount是把相同的内容换了一个挂载点的挂载方法
	return unix.Mount(config.Rootfs, config.Rootfs, "bind", unix.MS_BIND|unix.MS_REC, "")
}

/**
挂载
*/
func MountToRootfs(m *configc.Mount, rootfs string) error {
	logrus.WithField("init", true).Infof("mounting %#v to rootfs...", m)
	const defaultMountFlags = unix.MS_NOEXEC | unix.MS_NOSUID | unix.MS_NODEV
	var (
		dest = m.Destination
	)
	if !strings.HasPrefix(dest, rootfs) {
		dest = filepath.Join(rootfs, dest)
	}

	_, err := os.Stat(dest)
	if err != nil {
		// 不存在，则创建
		if err := os.MkdirAll(dest, 0755); err != nil {
			return err
		}
	}
	return mount(m, rootfs)
}

/**
真正执行挂载
*/
func mount(m *configc.Mount, rootfs string) error {
	var (
		flags = m.Flags
		dest  = m.Destination
	)
	if util.CleanPath(m.Destination) == "/dev" {
		flags &= ^unix.MS_RDONLY
	}
	if !strings.HasPrefix(dest, rootfs) {
		dest = filepath.Join(rootfs, dest)
	}
	// mount -t device src dest
	if err := unix.Mount(m.Source, dest, m.Device, uintptr(flags), m.Data); err != nil {
		logrus.WithField("init", true).Errorf("mount failed, cause: %s", err.Error())
		return err
	}
	return nil
}

/**
将该mount置为read only
*/
func RemountReadonly(m *configc.Mount) error {
	var (
		dest  = m.Destination
		flags = m.Flags
	)
	flags |= unix.MS_REMOUNT | unix.MS_BIND | unix.MS_RDONLY
	if err := unix.Mount("", dest, "", uintptr(flags), ""); err != nil {
		return err
	}
	return nil
}

/**
将rootfs 置为read only
*/
func SetRootfsReadonly() error {
	return unix.Mount("/", "/", "bind", unix.MS_BIND|unix.MS_REMOUNT|unix.MS_RDONLY|unix.MS_REC, "")
}
