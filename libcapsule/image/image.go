package image

import "time"

type Image struct {
	Id         string
	LayerId    string
	CreateTime time.Time
	Size       int64
}

type ImageRunArgs struct {
	ImageId      string
	ContainerId  string
	Args         []string
	Env          []string
	Cwd          string
	Hostname     string
	Cpushare     uint64
	Memory       int64
	Annotations  map[string]string
	Network      string
	PortMappings []string
	Detach       bool
	Volumes      []string
	Links        []string
}
