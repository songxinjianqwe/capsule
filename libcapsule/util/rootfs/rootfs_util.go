package rootfs

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func PrepareRoot(config *configc.Config) error {
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
	// The alternative (classic) way to create a read-only bind mount is to use the remount operation, for example:
	//      mount --bind olddir newdir
	//      mount -o remount,bind,ro olddir newdir
	// Note that a read-only bind will create a read-only mountpoint (VFS entry), but the original filesystem superblock will still be writable, meaning that the olddir will be writable, but the newdir will be read-only.
	// It's also possible to change nosuid, nodev, noexec, noatime, nodiratime and relatime VFS entry flags by "remount,bind" operation. It's impossible to change mount options recursively (for example with -o rbind,ro).
	// 另一方面，还可以阻止文件被移动或者链接
	// It creates a boundary that files cannot be moved or linked across
	return unix.Mount(config.Rootfs, config.Rootfs, "bind", unix.MS_BIND|unix.MS_REC, "")
}

/**
挂载
*/
func MountToRootfs(m *configc.Mount, rootfs string) error {
	logrus.WithField("init", true).Infof("mount %#v to rootfs...", m)
	var (
		dest = m.Destination
	)
	if !strings.HasPrefix(dest, rootfs) {
		dest = filepath.Join(rootfs, dest)
	}
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	return mount(m)
}

/**
真正执行挂载
*/
func mount(m *configc.Mount) error {
	var (
		flags = m.Flags
	)
	if util.CleanPath(m.Destination) == "/dev" {
		flags &= ^unix.MS_RDONLY
	}
	if err := unix.Mount(m.Source, m.Destination, m.Device, uintptr(flags), m.Data); err != nil {
		return err
	}
	return nil
}

func RemountReadonly(m *configc.Mount) error {
	var (
		dest  = m.Destination
		flags = m.Flags
	)
	for i := 0; i < 5; i++ {
		// There is a special case in the kernel for
		// MS_REMOUNT | MS_BIND, which allows us to change only the
		// flags even as an unprivileged user (i.e. user namespace)
		// assuming we don't drop any security related flags (nodev,
		// nosuid, etc.). So, let's use that case so that we can do
		// this re-mount without failing in a userns.
		flags |= unix.MS_REMOUNT | unix.MS_BIND | unix.MS_RDONLY
		if err := unix.Mount("", dest, "", uintptr(flags), ""); err != nil {
			switch err {
			case unix.EBUSY:
				time.Sleep(100 * time.Millisecond)
				continue
			default:
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("unable to mount %s as readonly max retries reached", dest)
}

func SetRootfsReadonly() error {
	return unix.Mount("/", "/", "bind", unix.MS_BIND|unix.MS_REMOUNT|unix.MS_RDONLY|unix.MS_REC, "")
}
