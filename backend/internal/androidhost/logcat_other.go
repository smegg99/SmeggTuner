//go:build !android || !cgo

// backend/internal/androidhost/logcat_other.go
package androidhost

// CaptureOutput does nothing where stdout and stderr already reach a terminal or a log.
func CaptureOutput() func() { return func() {} }
