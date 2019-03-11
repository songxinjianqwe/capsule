package constant

const (
	// 容器状态文件的文件名
	// 存放在 $RuntimeRoot/$containerId/下
	StateFilename       = "state.json"
	NotExecFlagFilename = "not-exec.flag"

	// 运行时文件的存放目录
	RuntimeRoot = "/var/run/capsule"
	// 各个容器的运行时文件的存放目录
	ContainerRuntimeRoot = RuntimeRoot + "/containers"
	// 容器配置文件，存放在运行capsule的cwd下
	ContainerConfigFilename = "config.json"
	// 容器Init进程的日志
	ContainerInitLogFilename = "container.log"
	// 容器Exec进程的日志名模板
	ContainerExecLogFilenamePattern = "exec-%s.log"
	IPAMDefaultAllocatorPath        = RuntimeRoot + "/network/ipam/subnet.json"

	// 重新执行本应用的command，相当于 重新执行./capsule
	ContainerInitCmd = "/proc/self/exe"
	// 运行容器init进程的命令
	ContainerInitArgs = "init"
)
