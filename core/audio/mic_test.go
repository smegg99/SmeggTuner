// core/audio/mic_test.go
package audio

import (
	"context"
	"os"
	"testing"
	"time"
)

// Real-hardware tests run only with SMEGGTUNER_MIC_TESTS=1 so headless runs stay green.
func TestDevicesEnumerate(t *testing.T) {
	if os.Getenv("SMEGGTUNER_MIC_TESTS") == "" {
		t.Skip("set SMEGGTUNER_MIC_TESTS=1 to run hardware tests")
	}
	devs, err := Devices()
	if err != nil {
		t.Fatal(err)
	}
	if len(devs) == 0 {
		t.Fatal("no capture devices")
	}
	for _, d := range devs {
		if d.ID == "" || d.Name == "" {
			t.Fatalf("bad device entry: %+v", d)
		}
	}
}

func TestMicCaptures(t *testing.T) {
	if os.Getenv("SMEGGTUNER_MIC_TESTS") == "" {
		t.Skip("set SMEGGTUNER_MIC_TESTS=1 to run hardware tests")
	}
	src, err := NewMicSource("")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ch, err := src.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	var total int
	for b := range ch {
		if b.SampleRate <= 0 {
			t.Fatalf("bad rate %d", b.SampleRate)
		}
		total += len(b.Samples)
	}
	_ = src.Stop()
	if total == 0 {
		t.Fatal("captured nothing")
	}
}
