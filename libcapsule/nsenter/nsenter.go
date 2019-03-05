package nsenter

/*
#include <stdlib.h>
#include "nsenter.h"
*/
import "C"
import (
	"github.com/sirupsen/logrus"
	"github.com/songxinjianqwe/capsule/libcapsule"
	"github.com/songxinjianqwe/capsule/libcapsule/util"
	"github.com/songxinjianqwe/capsule/libcapsule/util/exception"
	"os"
	"strconv"
	"strings"
	"unsafe"
)

/**
cgo is disabled when cross compiling.
build constraints exclude all Go files in xxx
*/
func init() {
	envInitializerType := os.Getenv(libcapsule.EnvInitializerType)
	if envInitializerType == "" {
		return
	}
	initializerType := libcapsule.InitializerType(envInitializerType)
	if initializerType != libcapsule.SetnsInitializer {
		return
	}
	logrus.WithField("init", true).Infof("got initializer type: %s", initializerType)
	configPipeEnv := os.Getenv(libcapsule.EnvConfigPipe)
	initPipeFd, err := strconv.Atoi(configPipeEnv)
	logrus.WithField("init", true).Infof("got config pipe env: %d", initPipeFd)
	if err != nil {
		panic(exception.NewGenericErrorWithContext(err, exception.SystemError, "converting EnvConfigPipe to int"))
	}
	// 读取config
	configPipe := os.NewFile(uintptr(initPipeFd), "configPipe")
	logrus.WithField("exec", true).Infof("open child pipe: %#v", configPipe)
	logrus.WithField("exec", true).Infof("starting to read namespaces from child pipe")

	// 先读出一个int，即namespaces的长度
	length, err := util.ReadIntFromFile(configPipe)
	if err != nil {
		logrus.WithField("exec", true).Errorf("read namespace length failed: %s", err.Error())
		panic(err)
	}
	// 再读出namespaces
	namespacesInBytes := make([]byte, length)
	if _, err := configPipe.Read(namespacesInBytes); err != nil {
		logrus.WithField("exec", true).Errorf("read namespaces failed: %s", err.Error())
		panic(exception.NewGenericErrorWithContext(err, exception.SystemError, "reading init config from configPipe"))
	}
	namespaces := strings.Split(string(namespacesInBytes), ",")
	logrus.WithField("exec", true).Infof("read namespaces: %v", namespaces)
	arg := make([]*C.char, 0) //C语言char*指针创建切片
	for i := range namespaces {
		char := C.CString(namespaces[i])
		// 循环结束后释放
		defer C.free(unsafe.Pointer(char))
		strPtr := (*C.char)(unsafe.Pointer(char))
		arg = append(arg, strPtr) //将char*指针加入到arg切片
	}
	status := int(C.nsenter((**C.char)(unsafe.Pointer(&arg[0])), C.int(len(namespaces))))
	if status < 0 {
		logrus.WithField("exec", true).Errorf("enter namespaces failed, status: %d", status)
		os.Exit(status)
	}
}
