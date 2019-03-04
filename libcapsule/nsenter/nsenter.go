package nsenter

/*
#cgo CFLAGS: -Wall
extern void nsenter();
void __attribute__((constructor)) init(void) {
	nsenter();
}
*/
import "C"
