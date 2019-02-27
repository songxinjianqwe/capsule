package cgroups

import "github.com/songxinjianqwe/rune/libcapsule/configc"

type CgroupData struct {
	root      string
	innerPath string
	config    *configc.CgroupConfig
	pid       int
}
