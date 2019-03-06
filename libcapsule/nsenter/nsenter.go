package nsenter

/*
#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <fcntl.h>
#include <errno.h>
#include <unistd.h>
const char* PRINT_PREFIX 	   		= "[EXEC]";
const char* ENV_CONFIG_PIPE      	= "_LIBCAPSULE_CONFIG_PIPE";
const char* ENV_INITIALIZER_TYPE 	= "_LIBCAPSULE_INITIALIZER_TYPE";
const char* NS_DELIMETER 			= ",";
const char*	EXEC_INITIALIZER  		= "exec";
const int ERROR 					= 1;

// __attribute__((constructor))：在main函数之前执行某个函数
// https://stackoverflow.com/questions/25704661/calling-setns-from-go-returns-einval-for-mnt-namespace
__attribute__((constructor)) void init(void) {
	const char* type = getenv(ENV_INITIALIZER_TYPE);
	if (!type || strcmp(type, EXEC_INITIALIZER) != 0) {
		return;
	}
	printf("%s start to read namespaces\n", PRINT_PREFIX);
	const char* config_pipe_env = getenv(ENV_CONFIG_PIPE);
	printf("%s read config pipe env: %s\n", PRINT_PREFIX, config_pipe_env);
	int config_pipe_fd = atoi(config_pipe_env);
	printf("%s config pipe fd: %d\n", PRINT_PREFIX, config_pipe_fd);
	if (config_pipe_fd <= 0) {
		printf("%s converting config pipe to int failed\n", PRINT_PREFIX);
		exit(ERROR);
	}
	// 读出长度
	char lenBuffer[4];
	if (read(config_pipe_fd, lenBuffer, 4) < 0) {
		printf("%s lenBuffer: %s\n", PRINT_PREFIX, lenBuffer);
		printf("%s read namespace length failed\n", PRINT_PREFIX);
		exit(ERROR);
	}
	// big endian
	int len = (lenBuffer[0] << 24) + (lenBuffer[1] << 16) + (lenBuffer[2] << 8) + lenBuffer[3];
	printf("%s read namespace len: %d\n", PRINT_PREFIX, len);

	// 再读出namespaces
	char namespaces[len];
	if (read(config_pipe_fd, namespaces, len) < 0) {
		printf("%s read namespaces failed\n", PRINT_PREFIX);
		exit(ERROR);
	}
	printf("%s read namespaces: %s\n", PRINT_PREFIX, namespaces);
	char* ns = strtok(namespaces, NS_DELIMETER);
	while(ns) {
		char* atPtr = strchr(ns, '@');
		if (atPtr) {
			*atPtr = '\0';
		}
        printf("%s current namespace_path is %s\n",PRINT_PREFIX, ns);
        int result = nsenter(ns);
        printf("\n");
        //if (result < 0) {
		//	exit(ERROR);
        //}
		ns = strtok(NULL, NS_DELIMETER);
	}
	printf("%s enter namespaces succeeded\n", PRINT_PREFIX);
}

int nsenter(char* namespace_path) {
    printf("%s entering namespace_path %s ...\n", PRINT_PREFIX, namespace_path);
    int fd = open(namespace_path, O_RDONLY);
    if (fd < 0) {
        printf("%s open %s failed, cause: %s\n", PRINT_PREFIX, namespace_path, strerror(errno));
        return -1;
    }
    printf("%s open %s, got fd: %d \n", PRINT_PREFIX, namespace_path, fd);
    // int setns(int fd, int nstype)
    // 参数fd表示我们要加入的namespace的文件描述符,它是一个指向/proc/[pid]/ns目录的文件描述符，可以通过直接打开该目录下的链接或者打开一个挂载了该目录下链接的文件得到。
    // 参数nstype让调用者可以去检查fd指向的namespace类型是否符合我们实际的要求。如果填0表示不检查。
    if (setns(fd, 0) < 0) {
        close(fd);
        // Linux中系统调用的错误都存储于 errno中，errno由操作系统维护，存储就近发生的错误，即下一次的错误码会覆盖掉上一次的错误。
        // 字符串显示错误信息 / strerror
        printf("%s enter namespace %s failed, cause: %s\n", PRINT_PREFIX, namespace_path, strerror(errno));
        return -1;
    } else {
        close(fd);
        printf("%s enter namespace %s succeeded\n", PRINT_PREFIX, namespace_path);
    	return 0;
    }
}

*/
import "C"
