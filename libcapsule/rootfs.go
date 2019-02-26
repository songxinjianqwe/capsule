package libcapsule

import (
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"strings"
)

/**
挂载
*/
func mountToRootfs(m *configc.Mount, rootfs string) error {
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
