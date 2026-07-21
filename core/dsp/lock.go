package dsp

import "math"

// LockTracker reports a stable lock once the reading has held still for `need` seconds: same reed
// count, each frequency within epsilon Hz of the last. A duration, not a count of observations, so
// how often the fine stage runs cannot leak into how long the technician must hold a note.
type LockTracker struct {
	need    float64 // seconds the reading must hold
	epsilon float64
	prev    []float64
	held    float64
}

// NewLockTracker takes the time a reading must hold still, in seconds.
func NewLockTracker(needSeconds, epsilonHz float64) *LockTracker {
	return &LockTracker{need: needSeconds, epsilon: epsilonHz}
}

func (l *LockTracker) Reset() {
	l.prev = nil
	l.held = 0
}

// Progress reports how far the reading is toward a lock, 0..1: held time over the hold it needs.
// A picture of the settle for the UI; it never gates anything and the lock decision stays in Observe.
func (l *LockTracker) Progress() float64 {
	if l.need <= 0 || l.held >= l.need {
		return 1
	}
	return l.held / l.need
}

// Configure changes the hold time and drift tolerance in place and restarts the settle clock. The
// engine calls it when the config changes under a running engine.
func (l *LockTracker) Configure(needSeconds, epsilonHz float64) {
	l.need = needSeconds
	l.epsilon = epsilonHz
	l.Reset()
}

// Observe folds in one fine result, dt seconds after the last one.
func (l *LockTracker) Observe(freqs []float64, dt float64) bool {
	stable := len(freqs) > 0 && len(freqs) == len(l.prev)
	if stable {
		for i := range freqs {
			if math.Abs(freqs[i]-l.prev[i]) > l.epsilon {
				stable = false
				break
			}
		}
	}
	l.prev = append(l.prev[:0], freqs...)

	if !stable {
		l.held = 0
		return false
	}
	l.held += dt
	return l.held >= l.need
}
