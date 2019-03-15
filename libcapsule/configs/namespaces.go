package configs

import (
	"fmt"
	"syscall"
)

type NamespaceType string

const (
	NEWNET NamespaceType = "NEWNET"
	NEWPID NamespaceType = "NEWPID"
	NEWNS  NamespaceType = "NEWNS"
	NEWUTS NamespaceType = "NEWUTS"
	NEWIPC NamespaceType = "NEWIPC"
)

func AllNamespaceTypes() []NamespaceType {
	return []NamespaceType{
		NEWIPC,
		NEWUTS,
		NEWNET,
		NEWPID,
		NEWNS,
	}
}

// Namespace defines configuration for each namespace.  It specifies an
// alternate path that is able to be joined via setns.
type Namespace struct {
	Type NamespaceType `json:"type"`
	Path string        `json:"path"`
}

type Namespaces []Namespace

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
	case NEWUTS:
		return "uts"
	}
	return ""
}

// NsFlag converts the namespace type to its flag
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
	case NEWUTS:
		return syscall.CLONE_NEWUTS
	}
	return 0
}

func (n *Namespaces) CloneFlags() uintptr {
	var flags uintptr
	for _, ns := range *n {
		flags |= ns.Type.NsFlag()
	}
	return flags
}

func (n *Namespace) GetPath(pid int) string {
	return fmt.Sprintf("/proc/%d/ns/%s", pid, n.Type.NsName())
}

func (n *Namespaces) Remove(t NamespaceType) bool {
	i := n.index(t)
	if i == -1 {
		return false
	}
	*n = append((*n)[:i], (*n)[i+1:]...)
	return true
}

func (n *Namespaces) Add(t NamespaceType, path string) {
	i := n.index(t)
	if i == -1 {
		*n = append(*n, Namespace{Type: t, Path: path})
		return
	}
	(*n)[i].Path = path
}

func (n *Namespaces) index(t NamespaceType) int {
	for i, ns := range *n {
		if ns.Type == t {
			return i
		}
	}
	return -1
}

func (n *Namespaces) Contains(t NamespaceType) bool {
	return n.index(t) != -1
}

func (n *Namespaces) PathOf(t NamespaceType) string {
	i := n.index(t)
	if i == -1 {
		return ""
	}
	return (*n)[i].Path
}
