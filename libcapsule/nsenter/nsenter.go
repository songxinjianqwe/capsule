package nsenter

/*
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <fcntl.h>
#include <errno.h>
#include <unistd.h>

const char* LOG_PREFIX 	   			= "[EXEC]";
const char* ENV_CONFIG_PIPE      	= "_LIBCAPSULE_CONFIG_PIPE";
const char* ENV_INITIALIZER_TYPE 	= "_LIBCAPSULE_INITIALIZER_TYPE";
const char* NS_DELIMETER 			= ",";
const char* CMD_DELIMETER 			= " ";
const char*	EXEC_INITIALIZER  		= "exec";
const int ERROR 					= 1;

void read_namespces_and_enter_them();
// __attribute__((constructor))：在main函数之前执行某个函数
// https://stackoverflow.com/questions/25704661/calling-setns-from-go-returns-einval-for-mnt-namespace
// https://lists.linux-foundation.org/pipermail/containers/2013-January/031565.html

__attribute__((constructor)) void init(void) {
	read_namespces_and_enter_them();
}

void read_namespces_and_enter_them() {
	const char* type = getenv(ENV_INITIALIZER_TYPE);
	if (!type || strcmp(type, EXEC_INITIALIZER) != 0) {
		return;
	}
	printf("%s start to read namespaces\n", LOG_PREFIX);
	const char* config_pipe_env = getenv(ENV_CONFIG_PIPE);
	printf("%s read config pipe env: %s\n", LOG_PREFIX, config_pipe_env);
	int config_pipe_fd = atoi(config_pipe_env);
	printf("%s config pipe fd: %d\n", LOG_PREFIX, config_pipe_fd);
	if (config_pipe_fd <= 0) {
		printf("%s converting config pipe to int failed\n", LOG_PREFIX);
		exit(ERROR);
	}
	// 读出长度
	char intBuffer[4];
	if (read(config_pipe_fd, intBuffer, 4) < 0) {
		printf("%s read namespace length failed\n", LOG_PREFIX);
		exit(ERROR);
	}

	// big endian
	int nsLen = (intBuffer[0] << 24) + (intBuffer[1] << 16) + (intBuffer[2] << 8) + intBuffer[3];
	printf("%s read namespace len: %d\n", LOG_PREFIX, nsLen);

	// 再读出namespaces
	char namespaces[nsLen];
	if (read(config_pipe_fd, namespaces, nsLen) < 0) {
		printf("%s read namespaces failed\n", LOG_PREFIX);
		exit(ERROR);
	}
	namespaces[nsLen] = '\0';
	printf("%s read namespaces: %s\n", LOG_PREFIX, namespaces);
	char* ns = strtok(namespaces, NS_DELIMETER);
	while(ns) {
        printf("%s current namespace_path is %s\n", LOG_PREFIX, ns);
        int result = nsenter(ns);
        printf("\n");
        if (result < 0) {
			exit(ERROR);
        }
		ns = strtok(NULL, NS_DELIMETER);
	}
	printf("%s enter namespaces succeeded\n", LOG_PREFIX);

	if (read(config_pipe_fd, intBuffer, 4) < 0) {
		printf("%s read cmd length failed\n", LOG_PREFIX);
		exit(ERROR);
	}

	int cmdLen = (intBuffer[0] << 24) + (intBuffer[1] << 16) + (intBuffer[2] << 8) + intBuffer[3];
	printf("%s read cmd len: %d\n", LOG_PREFIX, cmdLen);

	char cmd[cmdLen];
	if (read(config_pipe_fd, cmd, 1024) < 0) {
		printf("%s read cmd failed\n", LOG_PREFIX);
		exit(ERROR);
	}
	cmd[cmdLen] = '\0';
	printf("%s read cmd: %s\n", LOG_PREFIX, cmd);

	if (close(config_pipe_fd) < 0) {
		printf("%s close child pipe failed, cause: %s\n", LOG_PREFIX, strerror(errno));
	}
	int status = system(cmd);
	if (status < 0) {
		printf("%s system(%s) failed, cause: %s\n", LOG_PREFIX, cmd, strerror(errno));
	} else {
		printf("%s system(%s) succeeded\n", LOG_PREFIX, cmd);
	}
	exit(status);
}

int nsenter(char* namespace_path) {
    printf("%s entering namespace_path %s ...\n", LOG_PREFIX, namespace_path);
    int fd = open(namespace_path, O_RDONLY);
    if (fd < 0) {
        printf("%s open %s failed, cause: %s\n", LOG_PREFIX, namespace_path, strerror(errno));
        return -1;
    }
    // int setns(int fd, int nstype)
    // 参数fd表示我们要加入的namespace的文件描述符,它是一个指向/proc/[pid]/ns目录的文件描述符，可以通过直接打开该目录下的链接或者打开一个挂载了该目录下链接的文件得到。
    // 参数nstype让调用者可以去检查fd指向的namespace类型是否符合我们实际的要求。如果填0表示不检查。
    if (setns(fd, 0) < 0) {
        close(fd);
        // Linux中系统调用的错误都存储于 errno中，errno由操作系统维护，存储就近发生的错误，即下一次的错误码会覆盖掉上一次的错误。
        // 字符串显示错误信息 / strerror
        printf("%s enter namespace %s failed, cause: %s\n", LOG_PREFIX, namespace_path, strerror(errno));
        return -1;
    } else {
        close(fd);
        printf("%s enter namespace %s succeeded\n", LOG_PREFIX, namespace_path);
    	return 0;
    }
}

*/
import "C"
