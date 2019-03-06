package libcapsule

type Process struct {
	// The command to be run followed by any arguments.
	Args []string

	// Env specifies the environment variables for the config.
	Env []string

	// Cwd will change the processes current working directory inside the container's rootfs.
	Cwd string

	// Init specifies whether the config is the first config in the container.
	Init bool

	// Detach specifies the container is running frontend or backend
	Detach bool
	// if init = true, ID is container id, or ID is exec id.
	ID string
}
