package libcapsule

import (
	"io"
	"os"
)

type Process struct {
	// The command to be run followed by any arguments.
	Args []string

	// Env specifies the environment variables for the config.
	Env []string

	// User will set the uid and gid of the executing config running inside the container
	// local to the container's user and group configuration.
	User string

	// Cwd will change the processes current working directory inside the container's rootfs.
	Cwd string

	// Stdin is a pointer to a reader which provides the standard input stream.
	Stdin io.Reader

	// Stdout is a pointer to a writer which receives the standard output stream.
	Stdout io.Writer

	// Stderr is a pointer to a writer which receives the standard util stream.
	Stderr io.Writer

	// ExtraFiles specifies additional open files to be inherited by the container
	ExtraFiles []*os.File

	// Init specifies whether the config is the first config in the container.
	Init bool

	// Detach specifies the container is running frontend or backend
	Detach bool
}
