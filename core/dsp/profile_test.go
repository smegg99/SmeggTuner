package dsp

import (
	"math"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/audio"
	"smegg.me/smeggtuner/core/tuning"
)

// soloLConfig is a 16' rank pulled alone: the calibration sweep's setup, profiling on.
func soloLConfig() EngineConfig {
	return EngineConfig{
		A4:               440,
		ReedCount:        1,
		FineWindow:       3 * time.Second,
		Octaves:          []OctaveRequest{{Offset: -12, Reeds: 1}},
		ProfileHarmonics: true,
	}
}

// A solo rank in profiling mode: the reading must carry the reed's own partial ratios, which is
// what the calibration sweep records into its takes.
func TestProfilingMeasuresTheReedsOwnVoice(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 8 * time.Second,
		Reeds: []audio.ReedSpec{
			{Freq: 219, Amp: 0.40, Harmonics: []float64{0.30, 0, 0.12}},
		},
		NoiseAmp: 1e-4,
	}
	cap := runEngine(t, spec, soloLConfig())
	m := cap.lastFull()
	if len(m.Reeds) != 1 || len(m.Reeds[0].Harmonics) != 2 {
		t.Fatalf("want one reed with two partial ratios, got %+v", m.Reeds)
	}
	if r2 := m.Reeds[0].Harmonics[0]; math.Abs(r2-0.30) > 0.05 {
		t.Errorf("2nd partial ratio %.3f, the reed sounds 0.30", r2)
	}
	if r4 := m.Reeds[0].Harmonics[1]; math.Abs(r4-0.12) > 0.05 {
		t.Errorf("4th partial ratio %.3f, the reed sounds 0.12", r4)
	}
}

// lmProfiled is a 16'+8' register with the 16' rank's voice calibrated: at the A4 key the 16'
// sounds A3, so the profile entry keys on note 57.
func lmProfiled(r2 float64) EngineConfig {
	c := lmConfig()
	c.Profiles = []RankProfile{{Offset: -12, Note: tuning.Note(57), R2: r2}}
	return c
}

// The case phase cannot judge: an 8' tuned dead onto the 16's partial, zero beat, nothing rotating.
// The calibrated voice can - the line stands far over what the partial alone could be - so the rank
// is confirmed present. This is what the calibration sweep buys.
func TestProfileConfirmsARankAtZeroBeat(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 8 * time.Second,
		Reeds: []audio.ReedSpec{
			{Freq: 219, Amp: 0.40, Harmonics: []float64{0.25}},
			{Freq: 438, Amp: 0.45},
		},
		NoiseAmp: 1e-4,
	}
	cap := runEngine(t, spec, lmProfiled(0.25))
	m, ok := Aggregate(cap.all, 440, 0.5)
	if !ok {
		t.Fatal("no usable measurements")
	}
	if len(m.Reeds) != 2 {
		t.Fatalf("a coincident 8' must be confirmed by its amplitude, got %+v", m.Reeds)
	}
	if len(m.Bands) != 2 || m.Bands[1].GhostOnly {
		t.Errorf("the 8' band holds a confirmed voice, got %+v", m.Bands)
	}
}

// The same spectrum without the 8': the line matches the calibrated partial, so it stays a
// partial - even through the value dips a dying bellows causes, which used to need the phase lock.
func TestProfileRefusesABlockedRankByAmplitude(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 8 * time.Second,
		Reeds: []audio.ReedSpec{
			{Freq: 219, Amp: 0.40, Harmonics: []float64{0.25}},
		},
		NoiseAmp: 1e-4,
	}
	cap := runEngine(t, spec, lmProfiled(0.25))
	m, ok := Aggregate(cap.all, 440, 0.5)
	if !ok {
		t.Fatal("no usable measurements")
	}
	if len(m.Reeds) != 1 || m.Reeds[0].Octave != -12 {
		t.Fatalf("want only the 16' reed, got %+v", m.Reeds)
	}
	if len(m.Bands) != 2 || !m.Bands[1].GhostOnly {
		t.Errorf("the 8' band must report GhostOnly, got %+v", m.Bands)
	}
}
