package rootfs

import (
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"strings"
	"syscall"
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
	// 另一方面，还可以阻止文件被移动或者被链接
	// It creates a boundary that files cannot be moved or linked across

	// 可以让当前root的老root和新root不在同一个文件系统
	// bind mount是把相同的内容换了一个挂载点的挂载方法
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

/**
将当前root文件系统改为rootfs目录下的文件系统
把整个系统切换到一个新的root目录，而移除对之前root文件系统的依赖，这样就可以unmount原来的root文件系统
原来系统的mount信息都会消失！
并且ps命令返回的进程号也只有1号sh进程和ps -ef进程
而chroot是针对某个进程，系统的其他部分依旧运行于老的root目录中
*/
func PivotRoot(rootfs string) error {
	pivotDitName := ".pivot_root"
	pivotDir := filepath.Join(rootfs, pivotDitName)
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}
	// new root, put old
	// 老的root现在挂载在rootfs/.pivot_root上
	// 挂载点目前仍然可以在mount命令中看到
	if err := syscall.PivotRoot(rootfs, pivotDir); err != nil {
		return err
	}
	//切换到新的目录
	if err := syscall.Chdir("/"); err != nil {
		return err
	}
	pivotDir = filepath.Join("/", pivotDitName)
	// unmount
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return err
	}
	// 删除临时目录
	return os.Remove(pivotDir)
}
