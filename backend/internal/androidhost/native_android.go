//go:build android && cgo

// backend/internal/androidhost/native_android.go
package androidhost

/*
#include "native_android.h"
#include <stdlib.h>
*/
import "C"

import "unsafe"

func init() {
	host.send = func(id uint64, data []byte) bool {
		return C.tuner_send(C.jlong(id), unsafe.Pointer(unsafe.SliceData(data)), C.size_t(len(data))) != 0
	}
}

//export Java_com_wails_app_AndroidHost_nativeBind
func Java_com_wails_app_AndroidHost_nativeBind(env *C.JNIEnv, _ C.jclass, instance C.jobject) {
	C.tuner_bind(env, instance)
}

//export Java_com_wails_app_AndroidHost_nativeUnbind
func Java_com_wails_app_AndroidHost_nativeUnbind(env *C.JNIEnv, _ C.jclass, instance C.jobject) {
	C.tuner_unbind(env, instance)
}

//export Java_com_wails_app_AndroidHost_nativeComplete
func Java_com_wails_app_AndroidHost_nativeComplete(env *C.JNIEnv, _ C.jclass, id C.jlong, payload C.jbyteArray) {
	var size C.int
	data := C.tuner_bytes(env, payload, &size)
	defer C.free(data)
	host.complete(uint64(id), C.GoBytes(data, size))
}
