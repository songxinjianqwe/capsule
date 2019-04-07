package exception

// ErrorCode is the API util code type.
type ErrorCode int

// API util codes.
const (
	ContainerCreateOrRunError = iota
	ContainerIdEmptyError
	ContainerIdExistsError
	FactoryNewError
	ContainerNotExistsError
	ContainerLoadError
	ContainerNotRunningError
	ContainerStillRunningError
	ParentProcessSignalError
	ParentProcessCreateError
	ParentProcessStartError
	ParentProcessWaitError
	ContainerRootCreateError
	PipeError
	EnvError
	InitializerCreateError
	InitializerRunError
	PrepareRootError
	MountError
	SyscallExecuteCmdError
	SignalError
	SysctlError
	LookPathError
	HostnameError
	RootfsError
	CgroupsError
	CmdStartError
	CmdWaitError
	// network
	NetworkError
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
	// image
	ImageServiceError
	ImageIdExistsError
	ImageIdNotExistsError
	ImageLoadError
	ImageCreateError
	ImageRepositoriesDumpError
	BundleCreateError
	SpecSaveError
	UnionFsError
	UnionFsMountError
	DnsError
	HostsError
)

func (c ErrorCode) String() string {
	switch c {
	case ContainerCreateOrRunError:
		return "container create or run error"
	case ContainerIdEmptyError:
		return "container id cant be empty"
	case FactoryNewError:
		return "new factory error"
	case ContainerNotExistsError:
		return "container not exists error"
	case ContainerStillRunningError:
		return "container still running error"
	case ContainerLoadError:
		return "load container error"
	case ContainerNotRunningError:
		return "container not running"
	case ParentProcessSignalError:
		return "send signal to parent process error"
	case ParentProcessCreateError:
		return "create parent process error"
	case ParentProcessStartError:
		return "start parent process error"
	case ParentProcessWaitError:
		return "wait parent process error"
	case ContainerRootCreateError:
		return "create container root error"
	case EnvError:
		return "environment variables error"
	case PipeError:
		return "pipe error"
	case InitializerCreateError:
		return "create initializer error"
	case InitializerRunError:
		return "run initializer error"
	case ContainerIdExistsError:
		return "container id exists error"
	case PrepareRootError:
		return "prepare root error"
	case MountError:
		return "mount error"
	case SyscallExecuteCmdError:
		return "execute command error"
	case SignalError:
		return "send signal to process error"
	case SysctlError:
		return "sysctl write error"
	case LookPathError:
		return "look path error"
	case HostnameError:
		return "set hostname error"
	case RootfsError:
		return "set up rootfs error"
	case CgroupsError:
		return "config cgroups error"
	case CmdStartError:
		return "start cmd error"
	case CmdWaitError:
		return "wait cmd error"
	// network
	case NetworkError:
		return "network error"
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
	case VethMoveToNetNsError:
		return "move veth to net ns error"
	case PortMappingsConfigError:
		return "config port mappings error"
	case RouteAddError:
		return "route add error"
	case EnterNetNsError:
		return "enter network namespace error"
	// image
	case ImageServiceError:
		return "image service error"
	case ImageIdExistsError:
		return "image id exists error"
	case ImageLoadError:
		return "load image error"
	case ImageCreateError:
		return "create image error"
	case ImageRepositoriesDumpError:
		return "image repositories dump error"
	case SpecSaveError:
		return "save spec error"
	case UnionFsError:
		return "union fs error"
	case ImageIdNotExistsError:
		return "image id not exists error"
	case UnionFsMountError:
		return "union fs mount error"
	case BundleCreateError:
		return "create bundle error"
	case DnsError:
		return "dns error"
	case HostsError:
		return "hosts error"
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
