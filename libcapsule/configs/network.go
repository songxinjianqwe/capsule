package configs

type EndpointConfig struct {
	// uuid
	ID string
	// bridge 或者 loopback
	NetworkDriver string
	NetworkName   string
	PortMappings  []string
}
