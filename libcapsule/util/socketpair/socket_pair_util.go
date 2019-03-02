package socketpair

import (
	"golang.org/x/sys/unix"
	"os"
)

// NewSocketPair returns a new unix socket pair
// 和管道和命名管道相比，socketpair有以下特点：
// 1. 全双工
// 2. 可用于任意两个进程之间的通信
func NewSocketPair(name string) (parent *os.File, child *os.File, err error) {
	fds, err := unix.Socketpair(unix.AF_LOCAL, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return nil, nil, err
	}
	return os.NewFile(uintptr(fds[1]), name+"-parent"), os.NewFile(uintptr(fds[0]), name+"-child"), nil
}
