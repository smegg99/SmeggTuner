package dsp

import (
	"time"

	"smegg.me/smeggtuner/core/tuning"
)

// EngineConfig is the whole of what the engine can be tuned by, and it is deliberately the engine's
// own type, depending on nothing outside core/dsp (only core/tuning). The application maps its config
// file onto this at the services boundary. These are the timings and capture behaviour a technician
// turns; the deeper calibration constants travel with the package (see thresholds.go).
type EngineConfig struct {
	A4         float64
	Transpose  int
	ReedCount  int
	ManualNote tuning.Note
	Scale      tuning.ScaleNaming
	Hum50      bool
	Hum60      bool
	Highpass   bool
	ClockPPM   float64

	// CalibSecs is how long the mic path spends learning the room before it trusts a reading; 0
	// disables it (the file path starts mid-note).
	CalibSecs float64

	// FineWindow is the analysis window length: the single highest-leverage knob. It sets the
	// frequency resolution, the width of a spectral lobe (and so how close two reeds may sit and still
	// be told apart), and the slowest beat the engine can honestly measure. Three seconds is the
	// balance the whole detector is calibrated around.
	FineWindow time.Duration

	// LockHold is how long a reading must hold still before the engine reports a stable lock.
	// LockEpsilonHz is how far any reed may drift, in Hz, between fine results and still count as the
	// same reading.
	LockHold      time.Duration
	LockEpsilonHz float64
}

// DefaultEngineConfig returns the engine's own defaults: the one place the out-of-the-box timings are stated.
func DefaultEngineConfig() EngineConfig {
	c := EngineConfig{}
	c.fill()
	return c
}

// fill defaults the zero-valued knobs. The values here are the engine's calibrated defaults and must
// stay exact: the golden fixtures measure against them.
func (c *EngineConfig) fill() {
	if c.A4 == 0 {
		c.A4 = 440
	}
	if c.ReedCount == 0 {
		c.ReedCount = 1
	}
	if c.FineWindow == 0 {
		c.FineWindow = 3 * time.Second
	}
	if c.LockHold == 0 {
		c.LockHold = 1250 * time.Millisecond
	}
	if c.LockEpsilonHz == 0 {
		c.LockEpsilonHz = 0.1
	}
}
