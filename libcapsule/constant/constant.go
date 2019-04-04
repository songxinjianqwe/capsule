package constant

const (
	// 容器状态文件的文件名
	// 存放在 $RuntimeRoot/$containerId/下
	StateFilename       = "state.json"
	NotExecFlagFilename = "not-exec.flag"

	// 运行时文件的存放目录
	DefaultRuntimeRoot = "/var/run/capsule"
	// 各个容器的运行时文件的存放目录
	ContainerDir              = "containers"
	ImageDir                  = "images"
	ImageRepositoriesFilename = "repositories.json"
	// 容器配置文件，存放在运行capsule的cwd下
	ContainerConfigFilename = "config.json"
	// 容器Init进程的日志
	ContainerInitLogFilename = "container.log"
	// 容器Exec进程的日志名模板
	ContainerExecLogFilenamePattern = "exec-%s.log"
	IPAMDefaultAllocatorPath        = "/network/ipam/subnet.json"

	// 重新执行本应用的command，相当于 重新执行./capsule
	ContainerInitCmd = "/proc/self/exe"
	// 运行容器init进程的命令
	ContainerInitArgs = "init"

	// 容器初始化相关的常量
	EnvConfigPipe      = "_LIBCAPSULE_CONFIG_PIPE"
	EnvInitializerType = "_LIBCAPSULE_INITIALIZER_TYPE"
	/*
		一个进程默认有三个文件描述符，stdin、stdout、stderr
		外带的文件描述符在这三个fd之后
	*/
	DefaultStdFdCount = 3
)
