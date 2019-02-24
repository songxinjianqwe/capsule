package libcapsule

import (
	"fmt"
	"github.com/mrunalp/fileutils"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"github.com/songxinjianqwe/rune/libcapsule/util"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func prepareRoot(config *configc.Config) error {
	logrus.WithField("init", true).Info("preparing root...")
	flag := unix.MS_SLAVE | unix.MS_REC
	if config.RootPropagation != 0 {
		flag = config.RootPropagation
	}
	logrus.WithField("init", true).Info("mounting / in \"\" fs...")
	if err := unix.Mount("", "/", "", uintptr(flag), ""); err != nil {
		return err
	}
	logrus.WithField("init", true).Infof("mounting %s in bind fs...", config.Rootfs)
	return unix.Mount(config.Rootfs, config.Rootfs, "bind", unix.MS_BIND|unix.MS_REC, "")
}

func mountToRootfs(m *configc.Mount, rootfs string) error {
	logrus.WithField("init", true).Infof("mount %v to rootfs...", m)
	var (
		dest = m.Destination
	)
	if !strings.HasPrefix(dest, rootfs) {
		dest = filepath.Join(rootfs, dest)
	}

	switch m.Device {
	case "proc", "sysfs", "mqueue":
		if err := os.MkdirAll(dest, 0755); err != nil {
			return err
		}
		return mountPropagate(m, rootfs)
	case "tmpfs":
		copyUp := m.Extensions&configc.EXT_COPYUP == configc.EXT_COPYUP
		tmpDir := ""
		stat, err := os.Stat(dest)
		if err != nil {
			if err := os.MkdirAll(dest, 0755); err != nil {
				return err
			}
		}
		if copyUp {
			tmpdir, err := prepareTmp("/tmp")
			if err != nil {
				return util.NewGenericErrorWithContext(err, util.SystemError, "tmpcopyup: failed to setup tmpdir")
			}
			defer cleanupTmp(tmpdir)
			tmpDir, err = ioutil.TempDir(tmpdir, "runetmpdir")
			if err != nil {
				return util.NewGenericErrorWithContext(err, util.SystemError, "tmpcopyup: failed to create tmpdir")
			}
			defer os.RemoveAll(tmpDir)
			m.Destination = tmpDir
		}
		if err := mountPropagate(m, rootfs); err != nil {
			return err
		}
		if copyUp {
			if err := fileutils.CopyDirectory(dest, tmpDir); err != nil {
				errMsg := fmt.Errorf("tmpcopyup: failed to copy %s to %s: %v", dest, tmpDir, err)
				if err1 := unix.Unmount(tmpDir, unix.MNT_DETACH); err1 != nil {
					return util.NewGenericErrorWithContext(err1, util.SystemError, fmt.Sprintf("tmpcopyup: %v: failed to unmount", errMsg))
				}
				return errMsg
			}
			if err := unix.Mount(tmpDir, dest, "", unix.MS_MOVE, ""); err != nil {
				errMsg := fmt.Errorf("tmpcopyup: failed to move mount %s to %s: %v", tmpDir, dest, err)
				if err1 := unix.Unmount(tmpDir, unix.MNT_DETACH); err1 != nil {
					return util.NewGenericErrorWithContext(err1, util.SystemError, fmt.Sprintf("tmpcopyup: %v: failed to unmount", errMsg))
				}
				return errMsg
			}
		}
		if stat != nil {
			if err = os.Chmod(dest, stat.Mode()); err != nil {
				return err
			}
		}
		return nil
	default:
		// ensure that the destination of the mount is resolved of symlinks at mount time because
		// any previous mounts can invalidate the next mount's destination.
		// this can happen when a user specifies mounts within other mounts to cause breakouts or other
		// evil stuff to try to escape the container's rootfs.
		dest := filepath.Join(rootfs, m.Destination)
		if err := checkMountDestination(rootfs, dest); err != nil {
			return err
		}
		// update the mount with the correct dest after symlinks are resolved.
		m.Destination = dest
		if err := os.MkdirAll(dest, 0755); err != nil {
			return err
		}
		return mountPropagate(m, rootfs)
	}
	return nil
}

// checkMountDestination checks to ensure that the mount destination is not over the top of /proc.
// dest is required to be an abs path and have any symlinks resolved before calling this function.
func checkMountDestination(rootfs, dest string) error {
	invalidDestinations := []string{
		"/proc",
	}
	// White list, it should be sub directories of invalid destinations
	validDestinations := []string{
		// These entries can be bind mounted by files emulated by fuse,
		// so commands like top, free displays stats in container.
		"/proc/cpuinfo",
		"/proc/diskstats",
		"/proc/meminfo",
		"/proc/stat",
		"/proc/swaps",
		"/proc/uptime",
		"/proc/loadavg",
		"/proc/net/dev",
	}
	for _, valid := range validDestinations {
		path, err := filepath.Rel(filepath.Join(rootfs, valid), dest)
		if err != nil {
			return err
		}
		if path == "." {
			return nil
		}
	}
	for _, invalid := range invalidDestinations {
		path, err := filepath.Rel(filepath.Join(rootfs, invalid), dest)
		if err != nil {
			return err
		}
		if path != "." && !strings.HasPrefix(path, "..") {
			return fmt.Errorf("%q cannot be mounted because it is located inside %q", dest, invalid)
		}
	}
	return nil
}

// /tmp has to be mounted as private to allow MS_MOVE to work in all situations
func prepareTmp(topTmpDir string) (string, error) {
	tmpdir, err := ioutil.TempDir(topTmpDir, "runctop")
	if err != nil {
		return "", err
	}
	if err := unix.Mount(tmpdir, tmpdir, "bind", unix.MS_BIND, ""); err != nil {
		return "", err
	}
	if err := unix.Mount("", tmpdir, "", uintptr(unix.MS_PRIVATE), ""); err != nil {
		return "", err
	}
	return tmpdir, nil
}

func cleanupTmp(tmpdir string) error {
	unix.Unmount(tmpdir, 0)
	return os.RemoveAll(tmpdir)
}

func mountPropagate(m *configc.Mount, rootfs string) error {
	var (
		dest  = m.Destination
		flags = m.Flags
	)
	if util.CleanPath(dest) == "/dev" {
		flags &= ^unix.MS_RDONLY
	}

	copyUp := m.Extensions&configc.EXT_COPYUP == configc.EXT_COPYUP
	if !(copyUp || strings.HasPrefix(dest, rootfs)) {
		dest = filepath.Join(rootfs, dest)
	}

	if err := unix.Mount(m.Source, dest, m.Device, uintptr(flags), ""); err != nil {
		return err
	}

	for _, pflag := range m.PropagationFlags {
		if err := unix.Mount("", dest, "", uintptr(pflag), ""); err != nil {
			return err
		}
	}
	return nil
}
