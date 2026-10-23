//go:build android && cgo

// backend/internal/androidhost/logcat_android.go
package androidhost

/*
#cgo LDFLAGS: -llog
#include <android/log.h>
#include <stdlib.h>
*/
import "C"

import (
	"bufio"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// CaptureOutput sends stdout and stderr to logcat, which Android otherwise discards; the returned func flushes before exit.
func CaptureOutput() func() {
	r, w, err := os.Pipe()
	if err != nil {
		return func() {}
	}
	for _, fd := range []int{1, 2} {
		if err := syscall.Dup3(int(w.Fd()), fd, 0); err != nil {
			_ = r.Close()
			_ = w.Close()
			return func() {}
		}
	}
	tag := C.CString("SmeggTuner")
	done := make(chan struct{})
	go func() {
		defer close(done)
		in := bufio.NewReader(r)
		for {
			line, err := in.ReadString('\n')
			if line = strings.TrimRight(line, "\r\n"); line != "" {
				msg := C.CString(line)
				C.__android_log_write(C.int(priority(line)), tag, msg)
				C.free(unsafe.Pointer(msg))
			}
			if err != nil {
				return
			}
		}
	}()
	return func() {
		_ = w.Close()
		_ = syscall.Close(1)
		_ = syscall.Close(2)
		select {
		case <-done:
		case <-time.After(time.Second):
		}
	}
}
