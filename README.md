# capsule

<a name="aa54437d"></a>
# A Simplified OCI(Open Containers Initiative) Implementation, just like runC
[![Travis-CI](https://travis-ci.org/songxinjianqwe/capsule.svg)](_https://travis-ci.org/songxinjianqwe/capsule_)<br />[![GoDoc](https://godoc.org/github.com/songxinjianqwe/capsule?status.svg)](_http://godoc.org/github.com/songxinjianqwe/capsule_)<br />[![codecov](https://codecov.io/github/songxinjianqwe/capsule/coverage.svg)](_https://codecov.io/gh/songxinjianqwe/capsule_)<br />[![Report card](https://goreportcard.com/badge/github.com/songxinjianqwe/capsule)](_https://goreportcard.com/report/github.com/songxinjianqwe/capsule_)

[https://github.com/songxinjianqwe/capsule](_https://github.com/songxinjianqwe/capsule_)

<a name="ede5e031"></a>
# Project Structure
`Capsule`是一个CLI工具，提供了对容器的CRUD操作。<br />CLI与C-S架构主要的区别是CLI仅支持创建在本机中运行的容器，而C-S架构可以创建远程容器。<br />Docker是一个C-S架构的软件，而Docker底层依赖于runC，runC是实现了OCI标准的CLI软件，Docker以可执行程序的方式来调用runC管理容器。<br />`Capsule`实现了部分OCI标准，主体与runC类似，架构与实现方面部分参考了runC，尽可能简化其逻辑，只保留了容器核心技术的运用（如namespaces, cgroups, pivot root, network, image等)。<br />除了OCI标准外，`Capsule`也实现了容器网络与镜像管理的功能，这部分其实应该放在另一个软件中实现，但由于时间有限，暂时放到本项目中实现。

<a name="Features"></a>
# _Features_
由`Capsule`创建的容器可以提供一下功能：
* namespace 支持, 包括 uts, pid, mount, network，暂不支持user ns
* control group(linux cgroups) 支持，目前仅支持cpu与memory的控制
* 支持运行在用户提供的root fs上
* 容器网络, 包括容器间网络、容器与宿主机间网络、容器与外部网络
* 丰富的容器CLI命令支持, 包括 `list`, `state`, `create`, `run`, `start`, `kill`, `delete`, `exec`, `ps`, `log` and `spec`.
* 镜像管理，包括镜像导入(由Docker导出的镜像)，以类似于Docker CLI的方式运行容器（即不需要提供OCI标准的config.json）

<a name="Install"></a>
# Install
<a name="6d023f77"></a>
## Step0 go get "github.com/songxinjianqwe/capsule
<a name="3621df95"></a>
## Step1 开启宿主机的ip forward
所谓转发即当主机拥有多于一块的网卡时，其中一块收到数据包，根据数据包的目的ip地址将包发往本机另一网卡，该网卡根据路由表继续发送数据包。这通常就是路由器所要实现的功能。

bridge收到来自容器的请求时，根据数据包的目的IP(比如目的IP为公网IP，则匹配到默认路由default，默认路由到eth0)，将数据包转发到eth0，bridge和eth0不需要直连。

```
vi /usr/lib/sysctl.d/50-default.conf #命令（编辑配置文件）
net.ipv4.ip_forward=1               # 设置转发
sysctl –p
```
<a name="2d28d901"></a>
## Step2 安装iptables
CentOS7默认的防火墙不是iptables,而是firewalle.

```shell
#先检查是否安装了iptables
service iptables status
#安装iptables
yum install -y iptables
#安装iptables-services
yum install -y iptables-services

#停止firewalld服务
systemctl stop firewalld
#禁用firewalld服务
systemctl mask firewalld

#启用iptables
systemctl enable iptables.service
systemctl start iptables.service
systemctl status  iptables.service
```

```
#查看iptables现有规则
iptables -L -n

# 注意删掉REJCECT规则，否则在ping的时候会出现Destination Host Prohibited
# 比如说刚装好之后可能是这样的，注意把INPUT的第5条和FORWARD的第1条删掉
[root@localhost mycontainer]# iptables -L -n --line-number
Chain INPUT (policy ACCEPT)
num  target     prot opt source               destination
1    ACCEPT     all  --  0.0.0.0/0            0.0.0.0/0            state RELATED,ESTABLISHED
2    ACCEPT     icmp --  0.0.0.0/0            0.0.0.0/0
3    ACCEPT     all  --  0.0.0.0/0            0.0.0.0/0
4    ACCEPT     tcp  --  0.0.0.0/0            0.0.0.0/0            state NEW tcp dpt:22
5    REJECT     all  --  0.0.0.0/0            0.0.0.0/0            reject-with icmp-host-prohibited

Chain FORWARD (policy ACCEPT)
num  target     prot opt source               destination
1    REJECT     all  --  0.0.0.0/0            0.0.0.0/0            reject-with icmp-host-prohibited

Chain OUTPUT (policy ACCEPT)
num  target     prot opt source               destination
```


<a name="QuickStart"></a>
# QuickStart
<a name="22d5becf"></a>
## 以符合OCI规范的方式运行容器
首先需要了解一下OCI规范，此处可以参考[runC的README](https://github.com/opencontainers/runc)。<br />简单来说运行一个容器分为三步：
1. 准备一个rootfs，可以使用docker export导出
1. 准备一个config.json，以配置文件的方式来配置容器运行参数
1. 使用命令行工具

<a name="2e1aa009"></a>
### Step0 准备镜像
我们需要一个具有一些工具(如ifconfig, stress, gcc, iptables等)的镜像，这里提供一个示例Dockerfile来打镜像：

```dockerfile
FROM centos
ADD stress-1.0.4.tar.gz /tmp/
RUN yum install -y gcc automake autoconf libtool make net-tools.x86_64 iptables-services nmap-ncat.x86_64 && cd /tmp/stress-1.0.4 && ./configure && make && make install
```
使用`docker build`来构造镜像。<br />使用`docker export $(docker create $image_name) -o centos_with_utilities.tar`来导出镜像。<br />这样我们就得到了一个tar文件，这个文件解压后就是rootfs。
* 创建一个目录，作为容器的bundle目录。
* 将tar文件解压到该目录下的rootfs目录，`tar -xvf centos_with_utilities.tar -C $container_dir/rootfs/`
<a name="cdcd0eb1"></a>
### Step1 准备config.json
* 进入该目录，然后运行`capsule spec`，可以生成一个示例的config.json。

这里提供一个示例config.json，其中args就是容器运行的命令，比如这里就是sh：
```json
{
	"ociVersion": "1.0.1-dev",
	"process": {
		"user": {
			"uid": 0,
			"gid": 0
		},
		"args": [
			"sh"
		],
		"env": [
			"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
			"TERM=xterm"
		],
		"cwd": "/"
	},
	"root": {
		"path": "rootfs",
		"readonly": true
	},
	"hostname": "capsule",
	"mounts": [
		{
			"destination": "/proc",
			"type": "proc",
			"source": "proc"
		},
		{
			"destination": "/dev",
			"type": "tmpfs",
			"source": "tmpfs",
			"options": [
				"nosuid",
				"strictatime",
				"mode=755",
				"size=65536k"
			]
		},
		{
			"destination": "/dev/pts",
			"type": "devpts",
			"source": "devpts",
			"options": [
				"nosuid",
				"noexec",
				"newinstance",
				"ptmxmode=0666",
				"mode=0620",
				"gid=5"
			]
		},
		{
			"destination": "/dev/shm",
			"type": "tmpfs",
			"source": "shm",
			"options": [
				"nosuid",
				"noexec",
				"nodev",
				"mode=1777",
				"size=65536k"
			]
		},
		{
			"destination": "/dev/mqueue",
			"type": "mqueue",
			"source": "mqueue",
			"options": [
				"nosuid",
				"noexec",
				"nodev"
			]
		},
		{
			"destination": "/sys",
			"type": "sysfs",
			"source": "sysfs",
			"options": [
				"nosuid",
				"noexec",
				"nodev",
				"ro"
			]
		}
	],
	"linux": {
		"resources": {
			"devices": [
				{
					"allow": false,
					"access": "rwm"
				}
			],
			"memory": {
				"limit": 104857600
			},
			"cpu": {
				"shares": 512
			}
		},
		"namespaces": [
			{
				"type": "pid"
			},
			{
				"type": "uts"
			},
			{
				"type": "ipc"
			},
			{
				"type": "network"
			},
			{
				"type": "mount"
			}
		]
	}
}
```

<a name="6bad0ab0"></a>
### Step2 运行容器
保证当前目录下有一个config.json和一个rootfs目录(当然rootfs目录也可以放在别的地方，注意修改config.json中的root.path的值)。<br />在当前目录下运行`capsule run $container_name`，这样就可以运行起一个容器了。注意$container_name是一个唯一的id。

`capsule list`可以查看所有容器；<br />`capsule state $container_name`可以查看该容器的详细信息。

<a name="792ee4a8"></a>
## 以镜像的方式运行容器
<a name="2e1aa009-1"></a>
### Step0 准备镜像
这一步和上面一致，也是导出一个tar文件。
<a name="731a8817"></a>
### Step1 导入镜像
`capsule image create $image_name $tar_path`<br />将该镜像纳入到capsule管理，注意image_name也是一个唯一的id。<br />`capsule image list`可以查看所有镜像。
<a name="6bad0ab0-1"></a>
### Step2 运行容器
`capsule image run $image_name $args --name $container_name `<br />比如说`capsule image run centos sh --name centos_container`

<a name="Usage"></a>
# Usage
<a name="822b76a6"></a>
## 全局参数
capsule --root $root_dir<br />可以指定运行时文件的根目录，可选参数，默认值为 /var/run/capsule
<a name="create"></a>
## create
将有容器的config.json所在的目录称为bundle。<br />可以在bundle下使用`capsule create $container_name`来创建一个容器，容器会进入Created状态，也可以在任意目录，但要加入bundle参数，指明config.json的所在目录。<br />容器目前有三种状态，分别是：
* Created：在create命令执行后会进入的状态，容器的init process会阻塞在执行用户指定命令之前，等待start命令唤醒自己。
* Running：在start命令唤醒后会进入的状态，容器会执行用户指定命令。
* Stopped：容器启动失败或用户指定的命令执行完毕或被容器init process被kill后会进入的状态。

参数：

| Name | Short Name | Type | Usage | Default Value |
| --- | --- | --- | --- | --- |
| bundle  | b | string | path to the root of the bundle directory, defaults to the current directory | $cwd |
| network | net | string | network connected by container | capsule_bridge0(类似于docker0) |
| port | p | string array | port mappings, example: host port:container port | [] |


<a name="start"></a>
## start
`capsule start $container_name`可以启动一个Created状态的容器。<br />无参数，注意start的话默认情况下是前台运行的。
<a name="run"></a>
## run
run = create + start + destroy(对于前台运行的容器来说)<br />可以在`bundle`下使用`capsule run $container_name`来创建一个容器，容器会进入`Created`状态，也可以在任意目录，但要加入`bundle`参数，指明`config.json`的所在目录。<br />不指定-d或者-d false时容器为前台运行，当退出时容器随之退出并将自己销毁；指定-d时会以后台方式运行，可以使用`capsule list`或者`capsule state`来查看该容器状态。<br />参数：

| Name | Short Name | Type | Usage | Default Value |
| --- | --- | --- | --- | --- |
| bundle  | b | string | path to the root of the bundle directory, defaults to the current directory | $cwd |
| network | net | string | network connected by container | capsule_bridge0(类似于docker0) |
| port | p | string array | port mappings, example: host port:container port | [] |
| detach | d | bool | detach from the container's process | false |


<a name="list"></a>
## list
列出所有容器，已经被销毁的容器不会被显示。(Docker可以用`docker ps -a`来展示已经被销毁的容器，这里做了简化，已经被销毁的不再记录)。<br />示例：`capsule list`<br />ID                       PID         STATUS      IP            BUNDLE                                                      CREATED<br />capsule-demo-container   6995        Running     192.168.1.4   /var/run/capsule/images/containers/capsule-demo-container   2019-04-22T14:15:27.69530294-04:00<br />mysql                    6877        Running     192.168.1.3   /var/run/capsule/images/containers/mysql                    2019-04-22T14:07:31.989435108-04:00<br />redis                    2689        Running     192.168.1.2   /var/run/capsule/images/containers/redis                    2019-04-22T13:38:46.534200797-04:00
<a name="kill"></a>
## kill
可以对一个Created或Running状态的容器执行kill命令。<br />`capsule kill $container_name [$signal]`<br />这里$signal可以不填，默认是SIGTERM，也可以使用其他信号，如SIGKILL等。<br />其实就是对容器init process发送一个信号。
<a name="log"></a>
## log
可以查看一个容器的stdout和stderr日志。<br />`capsule log $container_name`<br />也可以查看某一次后台运行的exec的日志：`capsule log $container_name -exec $exec_id`<br />$exec_id是在exec -d执行后控制台打印出来的UUID。
<a name="ps"></a>
## ps
可以查看一个容器的进程信息，等同于`capsule exec $container_name ps `<br />`capsule ps $container_name`

<a name="delete"></a>
## delete
删除一个容器，如果想删除一个Created或Running的容器，需要加-f参数。<br />`capsule delete [-f] $container_name`<br />加-f相当于 `capsule kill $container_name SIGKILL` 再`capsule delete $container_name`
<a name="spec"></a>
## spec
在当前目录下生成一个示例spec，类似于下面的样子：<br />一般情况下只需要关心：
* args：目录
* env：环境变量
* hostname：主机名
* mounts：挂载
* cpu：linux.cpu.shares是容器所占用cpu的比例，默认为1024，即全部占用。
* memory：linux.memory.limit是容器最多使用的内存大小，单位是byte。
```json
{
	"ociVersion": "1.0.1-dev",
	"process": {
		"user": {
			"uid": 0,
			"gid": 0
		},
		"args": [
			"sleep", "24h"
		],
		"env": [
			"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
			"TERM=xterm"
		],
		"cwd": "/"
	},
	"root": {
		"path": "rootfs",
		"readonly": true
	},
	"hostname": "capsule",
	"mounts": [
		{
			"destination": "/proc",
			"type": "proc",
			"source": "proc"
		},
		{
			"destination": "/dev",
			"type": "tmpfs",
			"source": "tmpfs",
			"options": [
				"nosuid",
				"strictatime",
				"mode=755",
				"size=65536k"
			]
		},
		{
			"destination": "/dev/pts",
			"type": "devpts",
			"source": "devpts",
			"options": [
				"nosuid",
				"noexec",
				"newinstance",
				"ptmxmode=0666",
				"mode=0620",
				"gid=5"
			]
		},
		{
			"destination": "/dev/shm",
			"type": "tmpfs",
			"source": "shm",
			"options": [
				"nosuid",
				"noexec",
				"nodev",
				"mode=1777",
				"size=65536k"
			]
		},
		{
			"destination": "/dev/mqueue",
			"type": "mqueue",
			"source": "mqueue",
			"options": [
				"nosuid",
				"noexec",
				"nodev"
			]
		},
		{
			"destination": "/sys",
			"type": "sysfs",
			"source": "sysfs",
			"options": [
				"nosuid",
				"noexec",
				"nodev",
				"ro"
			]
		}
	],
	"linux": {
		"resources": {
			"devices": [
				{
					"allow": false,
					"access": "rwm"
				}
			],
			"memory": {
				"limit": 104857600
			},
			"cpu": {
				"shares": 512
			}
		},
		"namespaces": [
			{
				"type": "pid"
			},
			{
				"type": "uts"
			},
			{
				"type": "ipc"
			},
			{
				"type": "network"
			},
			{
				"type": "mount"
			}
		]
	}
}
```

<a name="state"></a>
## state
查看某个容器的信息<br />`capsule state $container_name [-d]`<br />如果希望查看详细信息，可以加-d参数，可以查看更为详细的信息。
<a name="exec"></a>
## exec
进入一个Created或Running的容器中执行命令。<br />`capsule exec $container_name $args [-e $env] [-cwd $cwd] [-d]`<br />指定-d可以以后台方式来运行此进程。
<a name="network"></a>
## network
network是一个二级命令，下面包含`create`, `delete`, `list`, `show`四个子命令。<br />网络通常会有一个driver参数，指定网络的驱动类型，理论上可以支持多种驱动，目前仅支持网桥，即bridge。
<a name="create-1"></a>
### create
创建一个网络，一般情况是创建一个指定网段的网桥。<br />`capsule network create $network_name -driver bridge -subnet $subnet`<br />subnet是一个网段，比如说192.168.1.0/24，在创建容器时可以使用-network $network_name来将该容器的IP地址的分配范围指定为该网络的网段。

<a name="delete-1"></a>
### delete
删除一个网络。<br />`capsule network delete $network_name -driver bridge`

<a name="list-1"></a>
### list
列出所有的网络，注意，如果没有创建任何网段，当第一次创建容器时会自动创建一个名为capsule_bridge0，网段为192.168.1.0/24的网桥，类似于Docker的docker0。<br />`capsule network list `
<a name="show"></a>
### show
显示一个网络的详细信息<br />`capsule network show $container_name`

<a name="image"></a>
## image
image同样是个二级目录，下面包含`create`, `delete`, `list`, `get`, `runc`, `destroyc`6个子命令。
<a name="create-2"></a>
### create
创建一个镜像<br />`capsule image create $image_name $tar_path`
<a name="delete-2"></a>
### delete
删除一个镜像<br />capsule image delete $image_name
<a name="list-2"></a>
### list
列出所有镜像<br />`capsule image list`
<a name="get"></a>
### get
显示一个镜像的信息。<br />`capsule image get $image_name`
<a name="runc"></a>
### runc
以镜像方式来启动一个容器，类似于Docker。<br />capsule image run $image_name command<br />-id $container_name<br />[-d]<br />[-workdir $workdir]<br />[-hostname $hostname]<br />[-env $k=$v]<br />[-cpushare $cpushare]<br />[-memory $memory_limit]<br />下面是spec里没有的,由capsule负责做的配置信息<br />[-link $container_name:$container_alias]<br />[-volume $host_dir/$container_dir:$host_dir]<br />[-network $network_name]<br />[-port $host_port:$container_host]<br />[-label $k=$v]

| Name | Short Name | Type | Usage | Default Value |
| --- | --- | --- | --- | --- |
| detach | d | bool | 是否以后台方式启动 | false |
| id |  | string | 容器名称，必填，唯一 |  |
| cwd |  | string | 容器启动后所处的工作目录 | / |
| env | e | string array | 环境变量 | [] |
| hostname | h | string | 主机名 | $container_name |
| cpushare | c | int64 | cpu比例 | 1024 |
| memory | m | uint64 | 最大内存 | 0，即无限制 |
| network | net | string | 网络名称 | capsule_bridge0 |
| port | p | string array | 端口映射，host_port:container_port | [] |
| label | l | string array | 容器标签 | [] |
| volume | v | string array | 数据卷，container_dir或者host_dir:container_dir | [] |
| link |  | string array | 容器间的连接，container_id:alias | [] |


<a name="destroyc"></a>
### destroyc
类似于capsule delete，实际上是capsule delete + 清理容器与镜像间的关联数据。<br />`capsule image destroyc $container_name [-f]`

<a name="f220d0cf"></a>
# 使用capsule来运行capsule-demo-app SpringBoot应用+MySQL+Redis
这里提到的capsule-demo-app参见[这个github仓库](https://github.com/songxinjianqwe/capsule-demo-app)。
<a name="2e1aa009-2"></a>
## Step0 准备镜像
首先我们需要在Docker中pull下mysql和redis镜像，然后使用docker export命令导出镜像为tar包。<br />capsule-demo-app的Dockerfile为：

```dockerfile
FROM java:8
VOLUME /tmp
ADD capsule-demo-app.jar app.jar
EXPOSE 8080
ENTRYPOINT [ "sh", "-c", "java -jar /app.jar"]
```
同样也要导出tar包，此时我们会有三个tar包。

<a name="731a8817-1"></a>
## Step1 导入镜像
`capsule image create $image_name $tar_path`<br />![image.png](https://cdn.nlark.com/yuque/0/2019/png/257642/1555989992423-5245f308-52c2-48e2-8e53-e767f2d3053a.png#align=left&display=inline&height=327&name=image.png&originHeight=802&originWidth=1830&size=1835383&status=done&width=746)
<a name="4b192b2d"></a>
## Step2 启动Redis
首先我们需要知道Dockerfile中有CMD或者ENTRYPOINT这样的语句用来指定启动时的命令，capsule为了简化没有做这一步，对capsule来说镜像==rootfs。启动命令需要自己输入。<br />通过[阅读Redis的Dockerfile](https://github.com/docker-library/redis/blob/dcc0a2a343ce499b78ca617987e8621e7d31515b/5.0/Dockerfile)，可以拿到启动命令，大概就是运行一个脚本，在同目录下可以读到这个docker-entrypoint.sh脚本代码。<br />因为capsule没有实现user namespace，容器中只能用root权限，所以我们需要手动修改脚本内容，将下图中红框部分的代码去掉，否则运行时会报错。<br />![image.png](https://cdn.nlark.com/yuque/0/2019/png/257642/1555990349815-b597db7b-efd0-49f3-bc83-9b3129ef80ee.png#align=left&display=inline&height=400&name=image.png&originHeight=400&originWidth=542&size=161853&status=done&width=542)

这里需要手动修改脚本，通过`capsule image list` 命令可以看到每个镜像对应的layer id。<br />在/var/run/capsule/images/layers/$layer_id下可以看到rootfs，然后修改该脚本文件：<br />![image.png](https://cdn.nlark.com/yuque/0/2019/png/257642/1555990690664-e7e528ed-cd2f-4ac3-9341-411d1d164017.png#align=left&display=inline&height=202&name=image.png&originHeight=493&originWidth=1822&size=849337&status=done&width=746)

然后使用`capsule image runc redis /usr/local/bin/docker-entrypoint.sh redis-server --id redis -p 6379:6379 -d`来启动redis容器。<br />我们分析一下这条命令：
* capsule image runc是根据镜像来启动容器的命令
* redis是镜像名
* /usr/local/bin/docker-entrypoint.sh redis-server是启动命令
* id即容器名，需要唯一，这里是redis
* p是port的缩写，指定端口映射，即将容器内的6379端口映射到宿主机的6379端口
* d是detach的缩写，指定后台运行

启动之后如果没有报错，则使用capsule image list命令来查看已经启动的容器。<br />如果STATUS是Running，则说明容器启动成功。<br />可以进入容器来使用redis-cli来检测是否真正OK。<br />capsule exec redis bash<br />redis-cli<br />127.0.0.1:6379> keys *<br />(empty list or set)<br />127.0.0.1:6379> set k1 v1<br />OK<br />127.0.0.1:6379> get k1<br />"v1"<br />127.0.0.1:6379> keys *<br />1) "k1"<br />127.0.0.1:6379> del k1<br />(integer) 1<br />127.0.0.1:6379> keys *<br />(empty list or set)<br />127.0.0.1:6379> exit<br />exit
<a name="33890a9a"></a>
## Step3 启动MySQL
类似于Redis，同样需要修改脚本文件。<br />将红框部分的代码删掉，否则启动时会报错error: exec: "/usr/local/bin/docker-entrypoint.sh": stat /usr/local/bin/docker-entrypoint.sh: permission denied。<br />![image.png](https://cdn.nlark.com/yuque/0/2019/png/257642/1555991866106-dbdb435c-daf3-4ffc-9103-d79a46454d1d.png#align=left&display=inline&height=511&name=image.png&originHeight=799&originWidth=1166&size=615707&status=done&width=746)<br />使用这条命令来启动mysql容器：`capsule image runc mysql "/usr/local/bin/docker-entrypoint.sh mysqld --user=root" -id=mysql -v /root/mysql/logs:/logs -v /root/mysql/data:/var/lib/mysql -p 3306:3306 -d`

我们分析一下这条命令：
* capsule image runc是根据镜像来启动容器的命令
* mysql是镜像名
* "/usr/local/bin/docker-entrypoint.sh mysqld --user=root"是启动命令，因为命令中也包含参数，所以用引号引起来，capsule中对于args数组长度为1的进行了特殊处理，如果包含空格则split后再赋值给args
* id即容器名，需要唯一，这里是mysql
* v是volume的缩写，指定volume可以使得容器在销毁后仍然在宿主机上保存部分文件，对于mysql这种需要持久化存储的应用来说volume是必要的，当然宿主机上的目录需要我们先行创建好。
* p是port的缩写，指定端口映射，即将容器内的6379端口映射到宿主机的6379端口
* d是detach的缩写，指定后台运行

启动之后我们需要进入容器中，创建一个名为demo的数据库schema，并且将外部访问权限由仅本机修改为任意host。<br />capsule exec mysql bash<br />mysql -uroot -p<br />密码为空，直接回车即可<br />> show databases;<br />> create database demo;<br />> use mysql;<br />> update user set host='%' where user='root';<br />> flush privileges;<br />> exit
<a name="6a2fd5c1"></a>
## Step4 启动Web应用
`capsule image runc capsule-demo-app "java -jar /app.jar" -id capsule-demo-container -e "SPRING_PROFILES_ACTIVE=prod" -p 8080:8080 -d -link mysql:mysql-container -link redis:redis-container`<br />这里使用link来指定连接的mysql和redis服务器。

如果遇到问题可以使用capsule log $container_name的方式来打印容器的stdout日志。

这个Web应用对外暴露了三个HTTP接口：

| HTTP Method | Path | Body | Description |
| --- | --- | --- | --- |
| GET | /users |  | 获得所有用户的信息 |
| GET | /users/$userId |  | 获得该用户的信息，会使用Redis缓存 |
| POST | /users | {<br />    "id": "tom",<br />    "nickName": "tom"<br />} | 添加一个用户 |


