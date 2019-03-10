package configs

type EndpointConfig struct {
	// uuid
	ID string `json:"id"`
	// bridge 或者 loopback
	NetworkDriver string   `json:"network_driver"`
	NetworkName   string   `json:"network_name"`
	PortMappings  []string `json:"port_mappings"`
}
