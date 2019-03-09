package configs

// Network defines configuration for a container's networking stack
//
// The network configuration can be omitted from a container causing the
// container to be setup with the host's networking stack
type Network struct {
	// Type sets the networks type, commonly veth and loopback
	Type string `json:"type"`

	// Name of the network interface
	Name string `json:"name"`

	// The bridge to use.
	Bridge string `json:"bridge"`

	// MacAddress contains the MAC address to set on the network interface
	MacAddress string `json:"mac_address"`

	// Address contains the IPv4 and mask to set on the network interface
	Address string `json:"address"`

	// Gateway sets the gateway address that is used as the default for the interface
	Gateway string `json:"gateway"`
}
