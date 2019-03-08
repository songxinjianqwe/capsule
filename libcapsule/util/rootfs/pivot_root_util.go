package rootfs

import (
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"syscall"
)

/*
将当前root文件系统改为rootfs目录下的文件系统
把整个系统切换到一个新的root目录，而移除对之前root文件系统的依赖，这样就可以unmount原来的root文件系统
原来系统的mount信息都会消失！
并且ps命令返回的进程号也只有1号sh进程和ps -ef进程
而chroot是针对某个进程，系统的其他部分依旧运行于老的root目录中
*/
func PivotRoot(rootfs string) error {
	logrus.Infof("pivot root...")
	pivotDitName := ".pivot_root"
	pivotDir := filepath.Join(rootfs, pivotDitName)
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}
	// new root, put old
	// 老的root现在挂载在rootfs/.pivot_root上
	// 挂载点目前仍然可以在mount命令中看到
	if err := syscall.PivotRoot(rootfs, pivotDir); err != nil {
		logrus.Errorf("pivot root failed, cause: %s", err.Error())
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
