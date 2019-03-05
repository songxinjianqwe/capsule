#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "nsenter.h"

const char* PRINT_PREFIX            = "[EXEC]";

int nsenter(char** namespaces, int len) {
    int i;
    for(i = 0; i < len; i++) {
        printf("%s namespaces[%d] is %s\n", PRINT_PREFIX, i, namespaces[i]);
    }
}

//void join_namespaces(char *nslist)
//{
//	int num = 0, i;
//	char *saveptr = NULL;
//	char *namespace = strtok_r(nslist, ",", &saveptr);
//	struct namespace_t {
//		int fd;
//		int ns;
//		char type[PATH_MAX];
//		char path[PATH_MAX];
//	} *namespaces = NULL;
//
//	if (!namespace || !strlen(namespace) || !strlen(nslist))
//		bail("ns paths are empty");
//
//	/*
//	 * We have to open the file descriptors first, since after
//	 * we join the mnt namespace we might no longer be able to
//	 * access the paths.
//	 */
//	do {
//		int fd;
//		char *path;
//		struct namespace_t *ns;
//
//		/* Resize the namespace array. */
//		namespaces = realloc(namespaces, ++num * sizeof(struct namespace_t));
//		if (!namespaces)
//			bail("failed to reallocate namespace array");
//		ns = &namespaces[num - 1];
//
//		/* Split 'ns:path'. */
//		path = strstr(namespace, ":");
//		if (!path)
//			bail("failed to parse %s", namespace);
//		*path++ = '\0';
//
//		fd = open(path, O_RDONLY);
//		if (fd < 0)
//			bail("failed to open %s", path);
//
//		ns->fd = fd;
//		ns->ns = nsflag(namespace);
//		strncpy(ns->path, path, PATH_MAX - 1);
//		ns->path[PATH_MAX - 1] = '\0';
//	} while ((namespace = strtok_r(NULL, ",", &saveptr)) != NULL);
//
//	for (i = 0; i < num; i++) {
//		struct namespace_t ns = namespaces[i];
//
//		if (setns(ns.fd, ns.ns) < 0)
//			bail("failed to setns to %s", ns.path);
//
//		close(ns.fd);
//	}
//
//	free(namespaces);
//}
