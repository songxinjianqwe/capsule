package exception

// ErrorCode is the API util code type.
type ErrorCode int

// API util codes.
const (
	ContainerCreateOrRunError = iota
	ContainerIdEmptyError
	FactoryNewError
	ContainerNotExistsError
	ContainerStateLoadFromDiskError
	ContainerNotRunning
	ParentProcessSignalError

	// network
	BridgeNetworkCreateError
	BridgeNetworkLoadError
	NetworkLinkNotFoundError
	NetworkLinkDeleteError
	InterfaceIPAndRouteSetError
	InterfaceSetUpError
	IPTablesSetError
	IPTablesDeleteError
	IPAMLoadError
	IPAMDumpError
	IPRunOutError
	IPReleaseError
	VethPairCreateError
	VethInitError
	VethMoveToNetNsError
	PortMappingsConfigError
	RouteAddError
	EnterNetNsError
)

func (c ErrorCode) String() string {
	switch c {
	case ContainerCreateOrRunError:
		return "container create or run error"
	case ContainerIdEmptyError:
		return "cantainer id cant be empty"
	case FactoryNewError:
		return "new factory error"
	case ContainerNotExistsError:
		return "container not exists"
	case ContainerStateLoadFromDiskError:
		return "load container state from disk error"
	case ContainerNotRunning:
		return "container not running"
	case ParentProcessSignalError:
		return "send signal to parent process error"
	// network
	case BridgeNetworkCreateError:
		return "bridge network create error"
	case InterfaceIPAndRouteSetError:
		return "set interface ip and route error"
	case InterfaceSetUpError:
		return "set interface up error"
	case IPTablesSetError:
		return "set iptables rule error"
	case IPTablesDeleteError:
		return "delete iptables rule error"
	case BridgeNetworkLoadError:
		return "load bridge network error"
	case NetworkLinkNotFoundError:
		return "network link not found error"
	case NetworkLinkDeleteError:
		return "network link delete error"
	case IPAMLoadError:
		return "ipam load error"
	case IPAMDumpError:
		return "ipam dump error"
	case IPRunOutError:
		return "ip run out error"
	case IPReleaseError:
		return "ip release error"
	case VethPairCreateError:
		return "create veth pair error"
	case VethInitError:
		return "init veth error"
	case PortMappingsConfigError:
		return "config port mappings error"
	case RouteAddError:
		return "route add error"
	case EnterNetNsError:
		return "enter network namespace error"
	default:
		return "unknown error"
	}
}

// Error is the API util type.
type Error interface {
	error
	// Returns the util code for this util.
	Code() ErrorCode
}
