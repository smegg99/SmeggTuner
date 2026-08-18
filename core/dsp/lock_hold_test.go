package dsp

import (
	"context"
	"math"
	"testing"
)

// A note played through a bellows swell: steady, then a stretch too uneven to read, then steady
// again. The lock is a duration of readings that held still, so the unread stretch may not count
// toward it - the engine measured nothing there. It used to: the tracker is handed the interval
// since the last reading, so the whole swell was banked as hold the moment the note settled, and
// the lock landed on the first reading after it.
func TestLockDoesNotCreditTheHopsItNeverSaw(t *testing.T) {
	const (
		sr    = 48000
		freq  = 440.0
		block = 4096 // sliceSource's, and the engine reads once per delivered block
	)
	// The swell: loud for less than the hold needs, a quarter as loud for long enough that no
	// window is even (steadyRatio), then loud again. Never quiet enough to lose the note, or the
	// tracker would simply reset and prove nothing.
	seg := func(seconds, amp float64, at int) []float32 {
		out := make([]float32, int(seconds*sr))
		for i := range out {
			out[i] = float32(amp * math.Sin(2*math.Pi*freq*float64(at+i)/sr))
		}
		return out
	}
	var samples []float32
	samples = append(samples, seg(2.5, 0.4, 0)...)
	samples = append(samples, seg(3, 0.1, len(samples))...)
	samples = append(samples, seg(6, 0.4, len(samples))...)

	cap := &capture{}
	if err := NewEngine(defaultCfg(), cap.emit).Run(context.Background(), &sliceSource{samples: samples, rate: sr}); err != nil {
		t.Fatal(err)
	}

	cap.mu.Lock()
	defer cap.mu.Unlock()

	readings, lockedAt := 0, -1
	for _, m := range cap.all {
		if len(m.Reeds) > 0 {
			readings++
		}
		if m.Locked {
			lockedAt = readings
			break
		}
	}
	if lockedAt < 0 {
		t.Fatal("never locked: the swell is meant to postpone the lock, not to cost it")
	}

	hold := DefaultEngineConfig().LockHold.Seconds()
	want := int(hold / (block / float64(sr)))
	if lockedAt < want {
		t.Errorf("locked on reading %d, want no sooner than %d: %.2f s of hold at %.0f ms a reading. "+
			"The swell produced no readings, so its seconds were credited without being watched",
			lockedAt, want, hold, block/float64(sr)*1000)
	}
}
