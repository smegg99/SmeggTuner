//go:build !js

// backend/common/emit/emit_desktop_test.go
package emit

import (
	"testing"
	"time"
)

func TestEmitWithoutAppIsNoop(t *testing.T) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		Emit("x", 1)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Emit blocked with no application running")
	}
}
