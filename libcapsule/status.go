package libcapsule

// ContainerStatus is the status of a container.
type ContainerStatus int

const (
	// Created is the status that denotes the container exists but has not been run yet.
	Created ContainerStatus = iota
	// Running is the status that denotes the container exists and is running.
	Running
	// Stopped is the status that denotes the container does not have a created or running config.
	Stopped
)

func (s ContainerStatus) String() string {
	switch s {
	case Created:
		return "Created"
	case Running:
		return "Running"
	case Stopped:
		return "Stopped"
	default:
		return "Unknown"
	}
}
