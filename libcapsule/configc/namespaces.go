package configc

import (
	"fmt"
	"syscall"
)

type NamespaceType string

const (
	NEWNET    NamespaceType = "NEWNET"
	NEWPID    NamespaceType = "NEWPID"
	NEWNS     NamespaceType = "NEWNS"
	NEWUTS    NamespaceType = "NEWUTS"
	NEWIPC    NamespaceType = "NEWIPC"
	NEWUSER   NamespaceType = "NEWUSER"
	NEWCGROUP NamespaceType = "NEWCGROUP"
)

// Namespace defines configuration for each namespace.  It specifies an
// alternate path that is able to be joined via setns.
type Namespace struct {
	Type NamespaceType `json:"type"`
	Path string        `json:"path"`
}

type Namespaces []Namespace

func (namespaces Namespaces) CloneFlags() uintptr {
	var flags uintptr = 1
	for _, ns := range namespaces {
		flags |= ns.Type.NsFlag()
	}
	return flags
}

// NsName converts the namespace type to its filename
func (ns NamespaceType) NsName() string {
	switch ns {
	case NEWNET:
		return "net"
	case NEWNS:
		return "mnt"
	case NEWPID:
		return "pid"
	case NEWIPC:
		return "ipc"
	case NEWUSER:
		return "user"
	case NEWUTS:
		return "uts"
	}
	return ""
}

// NsName converts the namespace type to its filename
func (ns NamespaceType) NsFlag() uintptr {
	switch ns {
	case NEWNET:
		return syscall.CLONE_NEWNET
	case NEWNS:
		return syscall.CLONE_NEWNS
	case NEWPID:
		return syscall.CLONE_NEWPID
	case NEWIPC:
		return syscall.CLONE_NEWIPC
	case NEWUSER:
		return syscall.CLONE_NEWUSER
	case NEWUTS:
		return syscall.CLONE_NEWUTS
	}
	return 0
}

func NamespaceTypes() []NamespaceType {
	return []NamespaceType{
		NEWUSER, // Keep user NS always first, don't move it.
		NEWIPC,
		NEWUTS,
		NEWNET,
		NEWPID,
		NEWNS,
	}
}

func (n *Namespace) GetPath(pid int) string {
	return fmt.Sprintf("/proc/%d/ns/%s", pid, n.Type.NsName())
}
