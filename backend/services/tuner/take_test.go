package tuner

import (
	"testing"
	"time"

	"smegg.me/smeggtuner/backend/core/dsp"
)

// takes is every analysis that wrote a take.
func (l *eventLog) takes() []dsp.Measurement {
	var out []dsp.Measurement
	for _, e := range l.all() {
		if e.name == EventMeasurement && e.meas.Taken {
			out = append(out, e.meas.Measurement)
		}
	}
	return out
}

// TakeNow is a no-op unless recording and while our own tone sounds. Recording set before Start is in
// force on the new run, and a request then takes the next analysis without waiting for a lock.
func TestTakeNow(t *testing.T) {
	initConfig(t)
	s := fixtureService(t)
	log := captureEvents(t, s)

	mustOK(t, s.Start())
	if !waitFor(6*time.Second, func() bool { return s.shown.Load() != 0 }) {
		t.Fatal("the engine never named a note")
	}
	s.TakeNow()
	time.Sleep(1500 * time.Millisecond)
	if n := len(log.takes()); n != 0 {
		t.Fatalf("%d takes while not recording, want 0", n)
	}
	mustOK(t, s.Stop())

	s.SetRecording(true)
	s.toneUntil.Store(time.Now().Add(time.Hour).UnixNano())
	mustOK(t, s.Start())
	if !waitFor(6*time.Second, func() bool { return s.shown.Load() != 0 }) {
		t.Fatal("the engine never named a note on the second run")
	}
	s.TakeNow()
	time.Sleep(600 * time.Millisecond)
	if n := len(log.takes()); n != 0 {
		t.Fatalf("%d takes while a tone sounds, want 0", n)
	}

	s.toneUntil.Store(0)
	s.TakeNow()
	if !waitFor(2500*time.Millisecond, func() bool { return len(log.takes()) > 0 }) {
		t.Fatal("a take request while recording never took")
	}
	if m := log.takes()[0]; m.Lock.Visible {
		t.Fatal("the take came from a lock, not from the request")
	}
}
