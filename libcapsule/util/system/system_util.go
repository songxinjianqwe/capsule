package system

import (
	"github.com/opencontainers/runc/libcontainer/system"
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/rune/libcapsule/cgroups"
	"github.com/songxinjianqwe/rune/libcapsule/configc"
	"golang.org/x/sys/unix"
	"os"
	"syscall"
	"unsafe"
)

// GetSubreaper returns the subreaper setting for the calling process
func GetSubreaper() (int, error) {
	var i uintptr

	if err := unix.Prctl(unix.PR_GET_CHILD_SUBREAPER, uintptr(unsafe.Pointer(&i)), 0, 0, 0); err != nil {
		return -1, err
	}

	return int(i), nil
}

const _P_PID = 1

type siginfo struct {
	si_signo int32
	si_errno int32
	si_code  int32
	// below here is a union; si_pid is the only field we use
	si_pid int32
	// Pad to 128 bytes as detailed in blockUntilWaitable
	pad [96]byte
}

// isWaitable returns true if the process has exited false otherwise.
// Its based off blockUntilWaitable in src/os/wait_waitid.go
func isWaitable(pid int) (bool, error) {
	si := &siginfo{}
	_, _, e := unix.Syscall6(unix.SYS_WAITID, _P_PID, uintptr(pid), uintptr(unsafe.Pointer(si)), unix.WEXITED|unix.WNOWAIT|unix.WNOHANG, 0, 0)
	if e != 0 {
		return false, os.NewSyscallError("waitid", e)
	}

	return si.si_pid != 0, nil
}

// isNoChildren returns true if err represents a unix.ECHILD (formerly syscall.ECHILD) false otherwise
func isNoChildren(err error) bool {
	switch err := err.(type) {
	case syscall.Errno:
		if err == unix.ECHILD {
			return true
		}
	case *os.SyscallError:
		if err.Err == unix.ECHILD {
			return true
		}
	}
	return false
}

// signalAllProcesses freezes then iterates over all the processes inside the
// manager's cgroups sending the signal s to them.
// If s is SIGKILL then it will wait for each process to exit.
// For all other signals it will check if the process is ready to report its
// exit status and only if it is will a wait be performed.
func SignalAllProcesses(m cgroups.CgroupManager, s os.Signal) error {
	var procs []*os.Process
	if err := m.Freeze(configc.Frozen); err != nil {
		logrus.Warn(err)
	}
	pids, err := m.GetAllPids()
	if err != nil {
		m.Freeze(configc.Thawed)
		return err
	}
	for _, pid := range pids {
		p, err := os.FindProcess(pid)
		if err != nil {
			logrus.Warn(err)
			continue
		}
		procs = append(procs, p)
		if err := p.Signal(s); err != nil {
			logrus.Warn(err)
		}
	}
	if err := m.Freeze(configc.Thawed); err != nil {
		logrus.Warn(err)
	}

	subreaper, err := system.GetSubreaper()
	if err != nil {
		// The error here means that PR_GET_CHILD_SUBREAPER is not
		// supported because this code might run on a kernel older
		// than 3.4. We don't want to throw an error in that case,
		// and we simplify things, considering there is no subreaper
		// set.
		subreaper = 0
	}

	for _, p := range procs {
		if s != unix.SIGKILL {
			if ok, err := isWaitable(p.Pid); err != nil {
				if !isNoChildren(err) {
					logrus.Warn("signalAllProcesses: ", p.Pid, err)
				}
				continue
			} else if !ok {
				// Not ready to report so don't wait
				continue
			}
		}

		// In case a subreaper has been setup, this code must not
		// wait for the process. Otherwise, we cannot be sure the
		// current process will be reaped by the subreaper, while
		// the subreaper might be waiting for this process in order
		// to retrieve its exit code.
		if subreaper == 0 {
			if _, err := p.Wait(); err != nil {
				if !isNoChildren(err) {
					logrus.Warn("wait: ", err)
				}
			}
		}
	}
	return nil
}
