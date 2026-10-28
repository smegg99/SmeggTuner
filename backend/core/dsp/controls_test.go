// backend/core/dsp/controls_test.go
package dsp

import (
	"context"
	"testing"
	"time"

	"smegg.me/smeggtuner/backend/core/audio"
)

func TestControlBurstPreservesUpdatesAndRecordingState(t *testing.T) {
	e := NewEngine(EngineConfig{}, func(Measurement) {})
	for range 80 {
		e.Update(func(c *EngineConfig) { c.A4 += 0.01 })
	}
	e.SetRecording(true)
	e.TakeNow()
	e.SetRecording(false)
	src := audio.NewSynthSource(audio.SynthSpec{SampleRate: 48000, Duration: time.Second})
	if err := e.Run(context.Background(), src); err != nil {
		t.Fatal(err)
	}
	if e.cfg.A4 < 440.79 || e.cfg.A4 > 440.81 {
		t.Fatalf("control updates dropped: A4=%v", e.cfg.A4)
	}
	if e.recording {
		t.Fatal("final controls lost: still recording")
	}
}
