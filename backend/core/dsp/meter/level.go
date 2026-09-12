package meter

import "math"

// LevelIntervalsPerSecond is the capture level statistics rate; one interval is half a second.
const LevelIntervalsPerSecond = 2

const (
	clipFloor = 500 // a flat top counts only at |x| >= 500 int16 units
	clipRun   = 8   // equal consecutive samples that latch clipping
)

// LevelInterval is one completed level interval of the int16 capture.
type LevelInterval struct {
	RMS     float32
	Peak    float32
	Clipped bool // a flat top: clipRun equal samples at |x| >= clipFloor
}

// LevelMeter is the capture's level accounting; an interval ends on the chunk that reaches it, the overshoot carries.
type LevelMeter struct {
	length     int
	frames     int
	sumSquares float64
	count      int
	peak       int32
	last       int16
	running    bool
	run        int
	clipped    bool
}

// NewLevelMeter sizes the interval for a nominal rate.
func NewLevelMeter(sampleRate int) LevelMeter {
	return LevelMeter{length: sampleRate / LevelIntervalsPerSecond}
}

// Reset is the capture ring reset: the open interval, its run and its latch start again.
func (m *LevelMeter) Reset() { *m = LevelMeter{length: m.length} }

// Add accumulates one chunk and returns the interval it completes; a latched interval stops tracking runs.
func (m *LevelMeter) Add(samples []int16) (LevelInterval, bool) {
	for _, x := range samples {
		v := int32(x)
		m.sumSquares += float64(v) * float64(v)
		m.count++
		if v < 0 {
			v = -v
		}
		m.peak = max(m.peak, v)
		switch {
		case m.clipped:
		case v < clipFloor:
			m.running, m.run = false, 0
		case m.running && x == m.last:
			m.run++
			m.clipped = m.run >= clipRun
		default:
			m.last, m.running, m.run = x, true, 1
		}
	}
	m.frames += len(samples)
	if m.frames < m.length {
		return LevelInterval{}, false
	}
	out := LevelInterval{Peak: float32(m.peak), Clipped: m.clipped}
	if m.count > 0 {
		out.RMS = float32(math.Sqrt(m.sumSquares / float64(m.count)))
	}
	*m = LevelMeter{length: m.length, frames: m.frames - m.length}
	return out, true
}
