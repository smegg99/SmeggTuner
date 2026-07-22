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

	// Octaves is the compound-register layout: one band per octave the register sounds, Offset in
	// semitones from the played key (-12 a 16', 0 the 8' ranks, +12 a 4'), Reeds the ranks sharing
	// that octave. Set only when the register spans octaves; empty (or offset-0 only) keeps the
	// single band of ReedCount, where the musette pair machinery lives. The service boundary maps
	// the pulled register's banks onto this.
	Octaves []OctaveRequest

	// ProfileHarmonics asks the engine to measure each reading's own partials (see
	// ReedMeasure.Harmonics). Set while a solo-rank register is pulled - a rank alone is the one
	// time a partial is unmistakably that reed's - and left off everywhere else.
	ProfileHarmonics bool

	// Profiles is what such measurements taught: per rank octave and sounding note, how loud the
	// rank's partials stand over its fundamental. The compound stage reads it to judge the
	// coincident case by amplitude (see thresholds.go, profileExcess). Empty means unprofiled, and
	// the phase rules stand alone.
	Profiles []RankProfile

	// LockHold is how long a reading must hold still before the engine reports a stable lock.
	// LockEpsilonHz is how far any reed may drift, in Hz, between fine results and still count as the
	// same reading.
	LockHold      time.Duration
	LockEpsilonHz float64
}

// RankProfile is one calibrated note of one rank, measured with the rank sounding alone. Note is
// the rank's own sounding note (the key shifted by its octave); R2 and R4 the second and fourth
// partial's amplitude over the fundamental's, zero where none was found.
type RankProfile struct {
	Offset int
	Note   tuning.Note
	R2, R4 float64
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
	// The compound stage walks bands lowest first and resolves the key against the lowest declared
	// rank; both need the layout ascending, so it is normalized once here.
	for i := 1; i < len(c.Octaves); i++ {
		for j := i; j > 0 && c.Octaves[j].Offset < c.Octaves[j-1].Offset; j-- {
			c.Octaves[j], c.Octaves[j-1] = c.Octaves[j-1], c.Octaves[j]
		}
	}
	for i := range c.Octaves {
		if c.Octaves[i].Reeds < 1 {
			c.Octaves[i].Reeds = 1
		}
	}
}
