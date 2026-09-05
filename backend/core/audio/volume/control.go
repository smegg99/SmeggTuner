// backend/core/audio/volume/control.go

// Package volume is the reference's automatic input volume: a noise floor learned from the
// capture's 0.5 s level intervals, and a control law that moves the OS capture level until a
// played note sits between 5 % and 27 % of full scale.
package volume

import "math"

const (
	fullScale = 32767

	floorMinRMS = 3.2767 // 0.01 % of full scale; quieter intervals leave the floor alone
	floorDown   = 0.15   // the floor follows a quieter interval fast
	floorQuiet  = 0.05   // and a quiet louder one slowly
	floorUp     = 0.001  // and anything else hardly
	quietRatio  = 10
	quietPeak   = 3932

	perSecond = 2    // level intervals per second
	weak      = 5.0  // percent of full scale under which a signal raises the level
	loud      = 27.0 // percent of full scale over which three intervals lower it
	minStep   = 0.002

	minLevel, maxLevel = 0.1, 1
)

// Interval is one completed 0.5 s level interval of the int16 capture.
type Interval struct {
	RMS     float32 // int16 units
	Peak    float32 // largest |sample|, int16 units
	Clipped bool    // a flat top latched in the interval
	Restart bool    // the capture ring restarted since the previous interval: the floor starts again
}

// Control is one capture's automatic volume state. The zero value is ready.
type Control struct {
	floor float32 // the noise floor N, int16 RMS units; 0 until learned
	clips int
	louds int
	hold  int
}

// Step runs one interval at the endpoint level 0..1 and returns the level to set and whether to set
// it. A muted endpoint learns nothing and is never moved.
func (c *Control) Step(iv Interval, muted bool, level float32) (float32, bool) {
	if iv.Restart {
		c.floor = 0
	}
	if muted {
		return level, false
	}
	c.learn(iv)

	r := iv.RMS / fullScale * 100
	signal := r >= 2*(c.floor/fullScale*100)
	if iv.Clipped {
		c.clips++
	} else {
		c.clips = 0
	}
	if c.hold > 0 {
		c.hold--
	}
	if signal && r > loud && !iv.Clipped {
		c.louds++
	} else {
		c.louds = 0
	}
	next := level
	switch {
	case c.clips >= 2:
		c.louds, c.hold = 0, int(0.5*perSecond)
		next = level - 0.2/perSecond
	case c.louds >= 3:
		next = level - 0.1/perSecond
	case c.hold == 0 && signal && r < weak:
		next = level + 0.05/perSecond
	}
	next = min(maxLevel, max(minLevel, next))
	return next, math.Abs(float64(next-level)) >= minStep
}

// learn folds one interval into the noise floor.
func (c *Control) learn(iv Interval) {
	if iv.RMS < floorMinRMS {
		return
	}
	n := c.floor
	quiet := !iv.Clipped && iv.RMS <= quietRatio*n && iv.Peak <= quietPeak
	switch {
	case n <= 0:
		n = iv.RMS
	case iv.RMS < n:
		n += floorDown * (iv.RMS - n)
	case quiet:
		n += floorQuiet * (iv.RMS - n)
	default:
		n += floorUp * (iv.RMS - n)
	}
	c.floor = max(0, n)
}
