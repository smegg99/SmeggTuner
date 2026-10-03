package meter

import "math"

// Warm-up and glitch constants of the reference session, in seconds and percent.
const (
	clockNominal     = 48000 // assumed when the device reports no rate
	clockWarmMin     = 1.0   // packets before this are not rated
	clockWarmFrom    = 2.0   // the latch is not tried before this
	clockWarmForce   = 20.0  // the latch is set regardless from here
	clockWarmWindow  = 1.0
	clockWarmCount   = 6
	clockWarmSpan    = 0.85
	clockWarmRange   = 0.1
	clockGlitchQuiet = 0.1 // a glitch ends this long after its last packet
)

// ClockStatus is where a clock calibration session stands.
type ClockStatus int

const (
	ClockPreparing   ClockStatus = iota // warming up, no reference point yet
	ClockCalibrating                    // measuring against the reference point
	ClockCalibrated                     // accepted; Value is the correction to store
	ClockUnreliable                     // timed out wide or hit a glitch; the correction to store is 0
)

// ClockReport is one estimator step.
type ClockReport struct {
	Status  ClockStatus
	Live    float64 // the rate difference since the reference point, percent; 0 until one is measured
	Spread  float64 // range of the estimates over the last 5 s, percent
	Value   float64 // the correction to store once Calibrated or Unreliable
	Timeout bool    // the 500 s limit ended the session
}

// Done reports a final status; the session ignores everything after it.
func (r ClockReport) Done() bool { return r.Status >= ClockCalibrated }

// ClockCalibration measures a capture device's frame rate against the host clock, as the reference's
// per-device calibration does. Deliberate difference: the rate since a start point is the least-squares
// fit over every packet since it, not the ratio of the newest packet to the start, because a capture
// callback's arrival time is far noisier than a driver timestamp.
type ClockCalibration struct {
	nominal float64
	report  ClockReport

	started bool
	t0, f0  float64 // the first packet
	warmFit clockFit
	warm    clockRing // estimates since the first packet, for the warm-up latch
	latched bool

	ref        bool
	tRef, fRef float64 // the reference point
	fit        clockFit
	tau        float64 // the newest packet's time since the reference point

	glitch                  bool
	glitchFirst, glitchLast float64

	accept clockAcceptance
}

// NewClockCalibration starts a session for a device of nominal frames per second.
func NewClockCalibration(nominal int) *ClockCalibration {
	if nominal <= 0 {
		nominal = clockNominal
	}
	return &ClockCalibration{nominal: float64(nominal)}
}

// Packet feeds one capture packet: device frames delivered up to its end and the host time it arrived,
// in seconds. A glitch packet (lost audio) is not measured and holds the glitch state.
func (c *ClockCalibration) Packet(frames int64, at float64, glitch bool) {
	if c.report.Done() {
		return
	}
	if glitch {
		if !c.glitch {
			c.glitch, c.glitchFirst = true, at
		}
		c.glitchLast = at
		return
	}
	if c.glitch && at-c.glitchLast > clockGlitchQuiet {
		c.skip(at - c.glitchFirst)
	}
	f := float64(frames)
	switch {
	case !c.started:
		c.started, c.t0, c.f0 = true, at, f
		c.warmFit.add(0, 0)
	case c.ref:
		c.fit.add(f-c.fRef, at-c.tRef)
		c.tau = at - c.tRef
	case c.latched:
		c.ref, c.tRef, c.fRef = true, at, f
		c.fit.add(0, 0)
	default:
		c.warmUp(f-c.f0, at-c.t0)
	}
}

// skip ends a glitch: its span leaves both clocks, the frames at the nominal rate.
func (c *ClockCalibration) skip(gap float64) {
	frames := math.Round(gap * c.nominal)
	if c.started {
		c.t0, c.f0 = c.t0+gap, c.f0+frames
	}
	if c.ref {
		c.tRef, c.fRef = c.tRef+gap, c.fRef+frames
	}
	c.glitch = false
}

// warmUp rates a packet against the first one and sets the latch once a second of rates agrees.
func (c *ClockCalibration) warmUp(frames, t float64) {
	c.warmFit.add(frames, t)
	rate := c.warmFit.rate()
	if t < clockWarmMin || frames <= 0 || rate <= 0 {
		return
	}
	c.warm.push(100*(rate-c.nominal)/c.nominal, t)
	switch {
	case t >= clockWarmForce:
		c.latched = true
	case t < clockWarmFrom:
	default:
		w := c.warm.window(t - clockWarmWindow)
		c.latched = w.n >= clockWarmCount && w.span >= clockWarmSpan && w.spread <= clockWarmRange
	}
}

// Update is the half-second estimator step, wall seconds after the session started.
func (c *ClockCalibration) Update(wall float64) ClockReport {
	if c.report.Done() {
		return c.report
	}
	if c.glitch {
		c.report = ClockReport{Status: ClockUnreliable, Live: c.report.Live, Spread: c.report.Spread}
		return c.report
	}
	if !c.ref {
		c.accept.hist.reset()
		c.report = ClockReport{Status: ClockPreparing}
		return c.report
	}
	c.report.Status = ClockCalibrating
	rate := float64(float32(c.fit.rate()))
	if rate <= 0 {
		return c.report
	}
	e := 100 * (rate - c.nominal) / c.nominal
	s := c.accept.step(e, c.tau, wall)
	c.report.Live, c.report.Spread = e, s.spread
	if !s.done {
		return c.report
	}
	c.report.Timeout = s.timeout
	if s.timeout && s.spread > clockTimeoutRange {
		c.report.Status = ClockUnreliable
		return c.report
	}
	c.report.Status, c.report.Value = ClockCalibrated, s.value
	return c.report
}

// clockFit is the least-squares line of host time over device frames: the time carries the callback
// jitter, the frames are exact. Running moments keep it exact over hours of packets.
type clockFit struct{ n, meanF, meanT, sFF, sFT float64 }

func (f *clockFit) add(frames, t float64) {
	f.n++
	dF := frames - f.meanF
	f.meanF += dF / f.n
	f.meanT += (t - f.meanT) / f.n
	f.sFF += dF * (frames - f.meanF)
	f.sFT += dF * (t - f.meanT)
}

// rate is frames per host second, 0 below two packets.
func (f *clockFit) rate() float64 {
	if f.n < 2 || f.sFT <= 0 {
		return 0
	}
	return f.sFF / f.sFT
}
