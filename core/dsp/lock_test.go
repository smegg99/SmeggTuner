package dsp

import "testing"

func TestLockTracker(t *testing.T) {
	const hop = 0.25

	l := NewLockTracker(0.5, 0.1)
	if l.Observe([]float64{440}, hop) || l.Observe([]float64{440.05}, hop) {
		t.Fatal("locked too early")
	}
	if !l.Observe([]float64{440.02}, hop) {
		t.Fatal("should lock once the reading has held for half a second")
	}
	if !l.Observe([]float64{440.03}, hop) {
		t.Fatal("stays locked while stable")
	}
	if l.Observe([]float64{441.0}, hop) {
		t.Fatal("jump must unlock")
	}
	l.Reset()
	if l.Observe([]float64{440}, hop) {
		t.Fatal("reset must clear history")
	}
}

// The point of holding a duration rather than a streak: measure three times as often and the lock
// still takes as long.
func TestLockTrackerIsRateIndependent(t *testing.T) {
	const need = 1.5

	for _, hop := range []float64{0.25, 0.085} {
		l := NewLockTracker(need, 0.1)

		var elapsed float64
		for range 100 {
			locked := l.Observe([]float64{440}, hop)
			elapsed += hop
			if locked {
				break
			}
		}

		// The first observation cannot lock, so the lock lands within one hop of the time asked for.
		if elapsed < need || elapsed > need+2*hop {
			t.Fatalf("hop %.3fs: locked after %.3fs, want about %.2fs", hop, elapsed, need)
		}
	}
}

func TestLockTrackerReedCountChange(t *testing.T) {
	l := NewLockTracker(0.25, 0.1)
	l.Observe([]float64{440, 442.6}, 0.25)
	if l.Observe([]float64{440}, 0.25) {
		t.Fatal("different reed count must not lock")
	}
}
