package image

import "github.com/opencontainers/runtime-spec/specs-go"

var defaultMounts = []specs.Mount{
	{
		Destination: "/proc",
		Type:        "proc",
		Source:      "proc",
		Options:     nil,
	},
	{
		Destination: "/dev",
		Type:        "tmpfs",
		Source:      "tmpfs",
		Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
	},
	{
		Destination: "/dev/pts",
		Type:        "devpts",
		Source:      "devpts",
		Options:     []string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620", "gid=5"},
	},
	{
		Destination: "/dev/shm",
		Type:        "tmpfs",
		Source:      "shm",
		Options:     []string{"nosuid", "noexec", "nodev", "mode=1777", "size=65536k"},
	},
	{
		Destination: "/dev/mqueue",
		Type:        "mqueue",
		Source:      "mqueue",
		Options:     []string{"nosuid", "noexec", "nodev"},
	},
	{
		Destination: "/sys",
		Type:        "sysfs",
		Source:      "sysfs",
		Options:     []string{"nosuid", "noexec", "nodev", "ro"},
	},
}

// 后面要加几个mount
func buildSpec(rootfsPath string, args []string, env []string, cwd string, hostname string, cpushare uint64, memory int64, annotations map[string]string, mounts []specs.Mount) *specs.Spec {
	env = append(env, "PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", "TERM=xterm")
	mounts = append(mounts, defaultMounts...)
	return &specs.Spec{
		Version: specs.Version,
		Root: &specs.Root{
			Path:     rootfsPath,
			Readonly: false,
		},
		Process: &specs.Process{
			Args: args,
			Env:  env,
			Cwd:  cwd,
		},
		Hostname:    hostname,
		Mounts:      mounts,
		Annotations: annotations,
		Linux: &specs.Linux{
			Resources: &specs.LinuxResources{
				Devices: []specs.LinuxDeviceCgroup{
					{
						Allow:  false,
						Access: "rwm",
					},
				},
				CPU: &specs.LinuxCPU{
					Shares: &cpushare,
				},
				Memory: &specs.LinuxMemory{
					Limit: &memory,
				},
			},
			Namespaces: []specs.LinuxNamespace{
				{
					Type: "pid",
				},
				{
					Type: "network",
				},
				{
					Type: "ipc",
				},
				{
					Type: "uts",
				},
				{
					Type: "mount",
				},
			},
		},
	}
}
