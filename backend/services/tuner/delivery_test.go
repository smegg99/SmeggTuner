// backend/services/tuner/delivery_test.go
package tuner

import (
	"context"
	"testing"
	"time"

	coreaudio "smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/tuning"
)

type intervalObserver struct{ measurements []dsp.Measurement }

func (o *intervalObserver) OnMeasurement(m dsp.Measurement) {
	o.measurements = append(o.measurements, m)
}

func observationRun(t *testing.T) (*run, *intervalObserver, func()) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	r := &run{ctx: ctx, events: make(chan dsp.Measurement, 1), observations: make(chan dsp.Measurement, 4)}
	observer := &intervalObserver{}
	svc := &Service{record: observer}
	done := make(chan struct{})
	go func() { defer close(done); r.observe(svc.observe) }()
	return r, observer, func() { close(r.observations); <-done }
}

func TestFullUIQueueCannotDiscardTakesOrReorderIntervals(t *testing.T) {
	r, observer, finish := observationRun(t)
	for i := 1; i <= 50; i++ {
		r.offer(dsp.Measurement{Taken: i%3 == 0, Note: tuning.Note(60 + i%12), Input: dsp.InputInterval{Seq: uint64(i)}})
	}
	finish()
	if len(observer.measurements) != 50 {
		t.Fatalf("observed %d frames", len(observer.measurements))
	}
	takes := 0
	for i, m := range observer.measurements {
		if m.Input.Seq != uint64(i+1) {
			t.Fatalf("interval reordered at %d: %d", i, m.Input.Seq)
		}
		if m.Taken {
			takes++
		}
	}
	if takes != 16 {
		t.Fatalf("takes=%d", takes)
	}
	if r.drops.Load() != 49 {
		t.Fatalf("UI drops=%d", r.drops.Load())
	}
}

func TestNoRecordingRunStillDeliversVolumeIntervals(t *testing.T) {
	r, observer, finish := observationRun(t)
	engine := dsp.NewEngine(dsp.EngineConfig{}, r.offer)
	src := coreaudio.NewSynthSource(coreaudio.SynthSpec{SampleRate: 48000, Duration: 2 * time.Second, Reeds: []coreaudio.ReedSpec{{Freq: 440, Amp: 0.2}}})
	if err := engine.Run(context.Background(), src); err != nil {
		t.Fatal(err)
	}
	finish()
	var seq uint64
	for _, m := range observer.measurements {
		if m.Taken {
			t.Fatal("unarmed run produced take")
		}
		if m.Input.Seq < seq {
			t.Fatalf("input sequence reversed: %d after %d", m.Input.Seq, seq)
		}
		seq = m.Input.Seq
	}
	if seq < 3 {
		t.Fatalf("volume observer got only %d intervals without recording", seq)
	}
	if r.drops.Load() == 0 {
		t.Fatal("test did not saturate UI queue")
	}
}
