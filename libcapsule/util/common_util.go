package util

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

func CleanPath(path string) string {
	// Deal with empty strings nicely.
	if path == "" {
		return ""
	}

	// Ensure that all paths are cleaned (especially problematic ones like
	// "/../../../../../" which can cause lots of issues).
	path = filepath.Clean(path)

	// If the path isn't absolute, we need to do more processing to fix paths
	// such as "../../../../<etc>/some/path". We also shouldn't convert absolute
	// paths to relative ones.
	if !filepath.IsAbs(path) {
		path = filepath.Clean(string(os.PathSeparator) + path)
		// This can't fail, as (by definition) all paths are relative to root.
		path, _ = filepath.Rel(string(os.PathSeparator), path)
	}

	// Clean the path again for good measure.
	return filepath.Clean(path)
}

func GetAnnotations(labels []string) (bundle string, userAnnotations map[string]string) {
	userAnnotations = make(map[string]string)
	for _, l := range labels {
		parts := strings.SplitN(l, "=", 2)
		if len(parts) < 2 {
			continue
		}
		if parts[0] == "bundle" {
			bundle = parts[1]
		} else {
			userAnnotations[parts[0]] = parts[1]
		}
	}
	return
}

func PrintSubsystemPids(subsystemName, cgroupName, context string, init bool) {
	bytes, err := ioutil.ReadFile(path.Join("/sys/fs/cgroup", subsystemName, cgroupName, "tasks"))
	if err != nil {
		logrus.Warnf("read pids failed, cause: %s", err.Error())
		return
	}
	if len(bytes) == 0 {
		if init {
			logrus.WithField("init", true).Warnf("[Pids of %s in %s] is EMPTY", cgroupName, subsystemName)
		} else {
			logrus.Warnf("[Pids of %s in %s] is EMPTY", cgroupName, subsystemName)
		}
		return
	}
	if init {
		logrus.WithField("init", true).Warnf("[Pids of %s in %s]%s, context is %s", cgroupName, subsystemName, string(bytes), context)
	} else {
		logrus.Warnf("[Pids of %s in %s]%s, context is %s", cgroupName, subsystemName, string(bytes), context)
	}
}

func WaitUserEnterGo() {
	scanner := bufio.NewScanner(os.Stdin)
	logrus.Warnf("【ATTENTION】Enter go to continue")
	scanner.Scan()
	ans := scanner.Text()
	for ans != "go" {
		logrus.Warnf("【ATTENTION】Enter go to continue")
		scanner.Scan()
		ans = scanner.Text()
	}
}

func SyncSignal(pid int, signal syscall.Signal) error {
	return unix.Kill(pid, signal)
}

func WaitSignal(sigs ...os.Signal) syscall.Signal {
	receivedChan := make(chan os.Signal, 1)
	signal.Notify(receivedChan, sigs...)
	received := <-receivedChan
	signal.Reset(sigs...)
	return received.(syscall.Signal)
}

func Int32ToBytes(n int32) ([]byte, error) {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	if err := binary.Write(bytesBuffer, binary.BigEndian, x); err != nil {
		return nil, err
	}
	return bytesBuffer.Bytes(), nil
}

func ReadIntFromFile(file *os.File) (int, error) {
	var x int32
	if err := binary.Read(file, binary.BigEndian, &x); err != nil {
		return 0, err
	}
	return int(x), nil
}

// NewSocketPair returns a new unix socket pair
func NewSocketPair(name string) (parent *os.File, child *os.File, err error) {
	fds, err := unix.Socketpair(unix.AF_LOCAL, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return nil, nil, err
	}
	return os.NewFile(uintptr(fds[1]), name+"-parent"), os.NewFile(uintptr(fds[0]), name+"-child"), nil
}
