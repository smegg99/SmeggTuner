package meter

import (
	"math"
	"testing"
)

// synthClock is a synthetic capture device: a packet every step host seconds carrying the frames its
// clock reached, arriving late by late(i) seconds.
type synthClock struct {
	nominal int
	step    float64
	frames  func(t float64) float64 // device frames at true host time t
	late    func(i int) float64
	glitch  func(i int) bool
}

// steady is a device running offset percent fast.
func steady(offset float64) func(float64) float64 {
	return func(t float64) float64 { return 48000 * (1 + offset/100) * t }
}

// run feeds c for seconds, stepping the estimator once per nominal half second of frames at the
// arrival time of the packet that completes it; it returns the last report and when it was taken.
func (d synthClock) run(c *ClockCalibration, seconds float64, each func(ClockReport, float64)) (ClockReport, float64) {
	half := float64(d.nominal) / 2
	next := half
	var r ClockReport
	var at float64
	for i := 0; ; i++ {
		t := float64(i+1) * d.step
		if t > seconds {
			return r, at
		}
		f := math.Floor(d.frames(t))
		arrive := t
		if d.late != nil {
			arrive += d.late(i)
		}
		c.Packet(int64(f), arrive, d.glitch != nil && d.glitch(i))
		for f >= next {
			next += half
			r, at = c.Update(arrive), arrive
			if each != nil {
				each(r, at)
			}
			if r.Done() {
				return r, at
			}
		}
	}
}

func TestClockCalibrationAcceptsAnExactClock(t *testing.T) {
	for _, offset := range []float64{0, 0.0123, -0.25} {
		d := synthClock{nominal: 48000, step: 0.01, frames: steady(offset)}
		c := NewClockCalibration(48000)
		var preparing, calibrating bool
		r, at := d.run(c, 100, func(r ClockReport, _ float64) {
			preparing = preparing || r.Status == ClockPreparing
			calibrating = calibrating || r.Status == ClockCalibrating
		})
		if r.Status != ClockCalibrated || r.Timeout || r.Value != round3(offset) {
			t.Fatalf("offset %v: %+v, want calibrated at %v", offset, r, round3(offset))
		}
		// Warm-up latches at 2 s, the next packet is the reference, acceptance needs tau >= 20.75 s.
		if at < 22.75 || at > 23.5 || !preparing || !calibrating {
			t.Fatalf("offset %v: accepted at %v s (preparing %v, calibrating %v), want 22.75..23.5", offset, at, preparing, calibrating)
		}
	}
}

// With fewer than six rated packets a second the warm-up never agrees, so the latch waits for 20 s.
func TestClockCalibrationForcesTheLatchAtTwentySeconds(t *testing.T) {
	d := synthClock{nominal: 48000, step: 0.5, frames: steady(0.004)}
	c := NewClockCalibration(48000)
	firstCalibrating := 0.0
	r, at := d.run(c, 100, func(r ClockReport, at float64) {
		if r.Status == ClockCalibrating && firstCalibrating == 0 {
			firstCalibrating = at
		}
	})
	if firstCalibrating < 20 || firstCalibrating > 21 {
		t.Fatalf("calibrating from %v s, want the forced latch at 20 s", firstCalibrating)
	}
	if r.Status != ClockCalibrated || r.Value != 0.004 || at < firstCalibrating+20.75 {
		t.Fatalf("%+v at %v s", r, at)
	}
}

// A glitch still open at an estimator step ends the session as unreliable with nothing to store.
func TestClockCalibrationGlitchAtAStepIsUnreliable(t *testing.T) {
	d := synthClock{nominal: 48000, step: 0.01, frames: steady(0.01), glitch: func(i int) bool { return i == 999 }}
	c := NewClockCalibration(48000)
	r, at := d.run(c, 100, nil)
	if r.Status != ClockUnreliable || r.Value != 0 || math.Abs(at-10) > 0.011 {
		t.Fatalf("%+v at %v s, want unreliable at the 10 s step", r, at)
	}
	c.Packet(48000*20, 20, false)
	if again := c.Update(20); again != r {
		t.Fatalf("a finished session moved on: %+v", again)
	}
}

// A glitch that closes between two steps is cut out of the measurement and the session carries on.
func TestClockCalibrationSkipsAClosedGlitch(t *testing.T) {
	lost := func(i int) bool { return i >= 1001 && i < 1011 } // 10.02..10.11 s, closed by 10.22 s
	d := synthClock{nominal: 48000, step: 0.01, frames: steady(0.02), glitch: lost}
	c := NewClockCalibration(48000)
	r, _ := d.run(c, 100, nil)
	if r.Status != ClockCalibrated || r.Value != 0.02 {
		t.Fatalf("%+v, want calibrated at 0.02", r)
	}
}

// A drifting clock never settles: at 500 s a narrow 5 s spread stores the current estimate.
func TestClockCalibrationTimeoutStoresTheCurrentEstimate(t *testing.T) {
	drift := func(t float64) float64 { return 48000 * (t + 3e-6*t*t/2) } // the rate rises 3 ppm a second
	d := synthClock{nominal: 48000, step: 0.05, frames: drift}
	c := NewClockCalibration(48000)
	r, at := d.run(c, 600, nil)
	if r.Status != ClockCalibrated || !r.Timeout || at < 500 || at > 500.6 {
		t.Fatalf("%+v at %v s, want calibrated by the 500 s timeout", r, at)
	}
	if r.Value != round3(r.Live) || r.Spread > clockTimeoutRange {
		t.Fatalf("stored %v for live %v, spread %v", r.Value, r.Live, r.Spread)
	}
}

// A clock that jumps 2 s just before the timeout is too wide over 5 s: unreliable, 0 stored.
func TestClockCalibrationWideTimeoutIsUnreliable(t *testing.T) {
	jump := func(t float64) float64 {
		f := 48000 * (t + 3e-6*t*t/2)
		if t > 495 {
			f += 2 * 48000
		}
		return f
	}
	d := synthClock{nominal: 48000, step: 0.05, frames: jump}
	r, at := d.run(NewClockCalibration(48000), 600, nil)
	if r.Status != ClockUnreliable || !r.Timeout || r.Value != 0 || r.Spread <= clockTimeoutRange || at < 500 {
		t.Fatalf("%+v at %v s, want unreliable by the timeout", r, at)
	}
}

func TestClockAcceptanceThresholds(t *testing.T) {
	feed := func(e func(tau float64) float64, until float64) clockStep {
		var a clockAcceptance
		var s clockStep
		for tau := 0.5; tau <= until && !s.done; tau += 0.5 {
			s = a.step(e(tau), tau, tau+2)
		}
		return s
	}
	flat := func(float64) float64 { return 0.0421 }
	if s := feed(flat, 20.5); s.done {
		t.Fatal("accepted before tau 20.75")
	}
	if s := feed(flat, 21); !s.done || s.value != 0.042 || s.timeout {
		t.Fatalf("flat history: %+v", s)
	}
	// 9 ppm of noise passes the 10 ppm range; a 6 ppm ramp over the window fails the 5 ppm drift.
	wobble := func(tau float64) float64 { return 0.0421 + 0.0009*2*math.Mod(tau, 1) }
	if s := feed(wobble, 21); !s.done {
		t.Fatal("9 ppm range refused")
	}
	ramp := func(tau float64) float64 { return 0.0421 + 0.0006*tau/20 }
	if s := feed(ramp, 40); s.done {
		t.Fatalf("a 6 ppm drift accepted: %+v", s)
	}
	wide := func(tau float64) float64 { return 0.0421 + 0.0011*2*math.Mod(tau, 1) }
	if s := feed(wide, 40); s.done {
		t.Fatal("an 11 ppm range accepted")
	}
}

func TestRound3TiesAwayFromZero(t *testing.T) {
	for _, c := range []struct{ in, want float64 }{{0.0125, 0.013}, {-0.0125, -0.013}, {0.01249, 0.012}, {2.0005, 2.001}} {
		if got := round3(c.in); got != c.want {
			t.Errorf("round3(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}
