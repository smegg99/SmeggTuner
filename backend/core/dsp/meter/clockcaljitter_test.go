package meter

import (
	"math"
	"math/rand/v2"
	"testing"
)

// lateness models for a capture callback: how long after its last frame a packet reaches us.
func uniformLate(j float64, seed uint64) func(int) float64 {
	r := rand.New(rand.NewPCG(seed, 1))
	return func(int) float64 { return j * r.Float64() }
}

func halfNormalLate(j float64, seed uint64) func(int) float64 {
	r := rand.New(rand.NewPCG(seed, 2))
	return func(int) float64 { return j * math.Abs(r.NormFloat64()) }
}

// stalls is 1 ms of lateness with one packet in a hundred held for 30 ms.
func stalls(seed uint64) func(int) float64 {
	r := rand.New(rand.NewPCG(seed, 3))
	return func(int) float64 {
		if r.IntN(100) == 0 {
			return 0.030
		}
		return 0.001 * r.Float64()
	}
}

// twoPoint drives the acceptance estimator with the reference's own time base: the newest packet
// against a reference point taken at refAt, with no warm-up.
func twoPoint(d synthClock, refAt, seconds float64) (clockStep, float64) {
	var a clockAcceptance
	nom := float64(d.nominal)
	next := nom / 2
	var fRef, tRef float64
	ref := false
	for i := 0; ; i++ {
		t := float64(i+1) * d.step
		if t > seconds {
			return clockStep{}, t
		}
		f := math.Floor(d.frames(t))
		arrive := t + d.late(i)
		if !ref && t >= refAt {
			ref, fRef, tRef = true, f, arrive
		}
		for f >= next {
			next += nom / 2
			if !ref || arrive <= tRef {
				continue
			}
			rate := float64(float32((f - fRef) / (arrive - tRef)))
			if s := a.step(100*(rate-nom)/nom, arrive-tRef, arrive); s.done {
				return s, arrive
			}
		}
	}
}

// Callback lateness of 1..5 ms against the reference thresholds. The reference's two-point ratio
// carries one packet's whole lateness divided by tau, so it needs 100 s to minutes to settle to
// 10 ppm and with 5 ms of half-normal lateness not at all inside 500 s. The least-squares time base
// averages the lateness over every packet: it settles within a minute, within two storage steps.
// Over 30 seeds per model: least squares at 28..60 s, error at most 13 ppm; the reference at
// 101..117 s (uniform 1 ms), 438..500 s (uniform 5 ms), 169..288 s (half-normal 1 ms), never
// (half-normal 5 ms), 96..138 s (1 ms with stalls).
func TestClockJitterNeedsTheLeastSquaresTimeBase(t *testing.T) {
	const offset = 0.0123
	models := []struct {
		name     string
		late     func(int) float64
		refFails bool // the reference time base does not accept before the timeout
	}{
		{"uniform 1 ms", uniformLate(0.001, 1), false},
		{"uniform 5 ms", uniformLate(0.005, 2), false},
		{"half-normal 1 ms", halfNormalLate(0.001, 3), false},
		{"half-normal 5 ms", halfNormalLate(0.005, 4), true},
		{"1 ms with 30 ms stalls", stalls(5), false},
	}
	for _, m := range models {
		d := synthClock{nominal: 48000, step: 0.01, frames: steady(offset), late: m.late}
		r, at := d.run(NewClockCalibration(48000), 600, nil)
		if r.Status != ClockCalibrated || r.Timeout || math.Abs(r.Value-offset) > 0.002 || at > 70 {
			t.Errorf("%s: least squares %+v at %.1f s, want a normal acceptance near %v within 70 s", m.name, r, at, offset)
		}

		s, refAt := twoPoint(d, 2, 600)
		accepted := s.done && !s.timeout
		t.Logf("%s: least squares stores %.3f %% at %.1f s; the reference time base accepted %v at %.1f s",
			m.name, r.Value, at, accepted, refAt)
		if accepted == m.refFails || (accepted && refAt < 90) {
			t.Errorf("%s: the reference time base accepted %v at %.1f s; the measurement this test documents changed", m.name, accepted, refAt)
		}
	}
}
