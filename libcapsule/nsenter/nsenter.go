package nsenter

/*
#include <stdio.h>
void nsexec();
// __attribute__((constructor))：在main函数之前执行某个函数
// https://stackoverflow.com/questions/25704661/calling-setns-from-go-returns-einval-for-mnt-namespace
// https://lists.linux-foundation.org/pipermail/containers/2013-January/031565.html
__attribute__((constructor)) void init(void) {
	nsexec();
}
*/
import "C"
