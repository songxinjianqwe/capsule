package libcapsule

import "os"

type ParentProcess interface {
	// pid returns the pid for the running process.
	pid() int

	// start starts the process execution.
	start() error

	// send a SIGKILL to the process and wait for the exit.
	terminate() error

	// wait waits on the process returning the process containerState.
	wait() (*os.ProcessState, error)

	// startTime returns the process start time.
	startTime() (uint64, error)

	signal(os.Signal) error

	externalDescriptors() []string

	setExternalDescriptors(fds []string)
}
