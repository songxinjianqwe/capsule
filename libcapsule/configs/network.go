package configs

type EndpointConfig struct {
	// uuid
	ID string
	// veth 或者 loopback
	Type         string
	PortMappings []string
}
