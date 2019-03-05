#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include "nsenter.h"

const char* PRINT_PREFIX  = "[EXEC]";

int nsenter(char** namespaces, int len) {
    int i;
    for(i = 0; i < len; i++) {
        printf("%s entering namespaces[%d] %s...\n", PRINT_PREFIX, i, namespaces[i]);
        int fd = open(namespaces[i], O_RDONLY);
        // int setns(int fd, int nstype)
        // 参数fd表示我们要加入的namespace的文件描述符,它是一个指向/proc/[pid]/ns目录的文件描述符，可以通过直接打开该目录下的链接或者打开一个挂载了该目录下链接的文件得到。
        // 参数nstype让调用者可以去检查fd指向的namespace类型是否符合我们实际的要求。如果填0表示不检查。
        if (setns(fd, 0) < 0) {
            close(fd);
            printf("%s enter namespace %s failed\n", PRINT_PREFIX, namespaces[i]);
            exit(1);
        } else {
            close(fd);
            printf("%s enter namespace %s succeeded\n", PRINT_PREFIX, namespaces[i]);
        }
    }
}