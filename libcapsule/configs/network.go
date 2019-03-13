package configs

type EndpointConfig struct {
	// uuid
	ID           string   `json:"id"`
	NetworkName  string   `json:"network_name"`
	PortMappings []string `json:"port_mappings"`
}
