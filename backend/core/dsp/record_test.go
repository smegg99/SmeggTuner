package dsp

import (
	"context"
	"testing"
	"time"

	"smegg.me/smeggtuner/backend/core/audio"
)

// runControlled runs a steady synthetic note with controls sent before Run, which applies them first.
func runControlled(t *testing.T, cfg EngineConfig, seconds float64, before func(*Engine), reeds ...float64) []Measurement {
	t.Helper()
	spec := audio.SynthSpec{SampleRate: 48000, Duration: time.Duration(seconds * float64(time.Second))}
	for _, f := range reeds {
		spec.Reeds = append(spec.Reeds, audio.ReedSpec{Freq: f, Amp: 0.2})
	}
	var out []Measurement
	e := NewEngine(cfg, func(m Measurement) { out = append(out, m) })
	before(e)
	if err := e.Run(context.Background(), audio.NewSynthSource(spec)); err != nil {
		t.Fatal(err)
	}
	return out
}

func takes(ms []Measurement) []int {
	var at []int
	for i, m := range ms {
		if m.Taken {
			at = append(at, i)
		}
	}
	return at
}

func TestEngineRecordsOneTakePerLock(t *testing.T) {
	ms := runControlled(t, EngineConfig{}, 10, func(e *Engine) { e.SetRecording(true) }, 440.5)
	at := takes(ms)
	if len(at) != 1 {
		t.Fatalf("%d takes over a held note", len(at))
	}
	m := ms[at[0]]
	if m.Frame != FrameMeasured || !m.Lock.Internal || !m.Lock.Recorded || len(m.Reeds) != 1 || !m.Reeds[0].Present {
		t.Fatalf("take %+v", m)
	}
	for _, later := range ms[at[0]:] {
		if !later.Lock.Recorded {
			t.Fatalf("the recorded mark must stay while the note does: %+v", later.Lock)
		}
	}
	if last := ms[len(ms)-1]; last.Lock.Internal {
		t.Fatal("the take restarts the history: the lock must drop")
	}
}

func TestEngineTakeNowTakesTheFirstMeasurement(t *testing.T) {
	ms := runControlled(t, EngineConfig{}, 4, func(e *Engine) { e.SetRecording(true); e.TakeNow() }, 440.5)
	at := takes(ms)
	if len(at) != 1 {
		t.Fatalf("takes at %v", at)
	}
	for _, m := range ms[:at[0]] {
		if m.Frame == FrameMeasured {
			t.Fatal("take now must take the first measured analysis")
		}
	}
	if ms[at[0]].Lock.Internal {
		t.Fatal("this take had no lock")
	}
}

func TestEngineTakesNothingWhenNotRecordingOrFrozen(t *testing.T) {
	for name, before := range map[string]func(*Engine){
		"not recording": func(e *Engine) { e.TakeNow() },
		"frozen":        func(e *Engine) { e.SetRecording(true); e.TakeNow(); e.Freeze(true) },
	} {
		if at := takes(runControlled(t, EngineConfig{}, 8, before, 440.5)); len(at) != 0 {
			t.Fatalf("%s: takes at %v", name, at)
		}
	}
}

func TestEngineBassSessionsRecord(t *testing.T) {
	cfg := EngineConfig{Bass: true, Octaves: []OctaveRequest{{0, 1}, {12, 1}, {24, 1}}}
	ms := runControlled(t, cfg, 4, func(e *Engine) { e.SetRecording(true); e.TakeNow() }, 110.1, 220.3, 440.5)
	// Take now writes the first measured analysis with all three octaves; a later lock on a corrected
	// base note may take again, as any recording does.
	at := takes(ms)
	if len(at) == 0 || len(ms[at[0]].Reeds) != 3 {
		t.Fatalf("bass takes at %v", at)
	}
	for _, m := range ms[:at[0]] {
		if m.Frame == FrameMeasured {
			t.Fatal("take now must take the first measured analysis")
		}
	}
}

func TestEngineStoppingDropsAPendingTakeNow(t *testing.T) {
	ms := runControlled(t, EngineConfig{}, 10, func(e *Engine) {
		e.SetRecording(true)
		e.TakeNow()
		e.SetRecording(false)
		e.SetRecording(true)
	}, 440.5)
	at := takes(ms)
	if len(at) != 1 || !ms[at[0]].Lock.Internal {
		t.Fatalf("takes at %v: the only take must wait for a lock", at)
	}
}
