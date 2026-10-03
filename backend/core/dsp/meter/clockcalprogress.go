package meter

import "math"

// Progress bar constants of the reference, percent of the bar and percent of rate.
const (
	clockProgressStart = 5
	clockProgressNear  = 0.0005 // a 5 s spread at or under 5 ppm counts as settled
	clockProgressHalf  = 50.0
	clockProgressRise  = 45.0 // gained over clockProgressRiseT once settled
	clockProgressTop   = 95
	clockProgressRiseT = 20.0
)

// ClockProgress is the reference's calibration progress bar, 0..100; NewClockProgress starts one.
type ClockProgress struct {
	shown  int
	widest float64 // the widest 5 s spread seen
	timed  bool
	since  float64 // session time the settled timer started
}

// NewClockProgress is the bar at a session start.
func NewClockProgress() ClockProgress { return ClockProgress{shown: clockProgressStart} }

// Next folds one estimator step taken wall seconds into the session into the bar and returns it.
// A shrinking spread fills the first half; once settled, a 20 s timer fills up to 95; it never falls.
func (p *ClockProgress) Next(r ClockReport, wall float64) int {
	switch r.Status {
	case ClockCalibrated:
		v := 0
		if p.widest > 0 {
			v = 100
		}
		*p = ClockProgress{shown: v}
		return v
	case ClockUnreliable:
		*p = ClockProgress{}
		return 0
	}
	s := r.Spread
	p.widest = max(p.widest, s)
	var v float64
	if s > clockProgressNear && p.widest > clockProgressNear {
		v = max(clockProgressStart, half(min(clockProgressHalf, max(0, clockProgressHalf*(p.widest-s)/(p.widest-clockProgressNear)))))
	} else {
		if !p.timed {
			p.timed, p.since = true, wall
		}
		v = half(min(clockProgressTop, max(clockProgressHalf, clockProgressHalf+clockProgressRise*(wall-p.since)/clockProgressRiseT)))
	}
	p.shown = min(clockProgressTop, max(0, max(p.shown, int(v))))
	return p.shown
}

// Shown is the bar as it stands.
func (p ClockProgress) Shown() int { return p.shown }

// Settled reports the second half: the spread held under 5 ppm and the 20 s timer runs.
func (p ClockProgress) Settled() bool { return p.timed }

// half rounds a nonnegative value half up.
func half(x float64) float64 { return math.Floor(x + 0.5) }
