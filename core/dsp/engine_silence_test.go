package dsp

import (
	"context"
	"math"
	"math/rand"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/audio"
)

// A recording has silence before and after the note, and detection is relative, so in the quiet noise
// clears the gate and the engine names a note nobody played. A real "H '8" recording showed A#2.
func TestEngineDoesNotHearNotesInSilence(t *testing.T) {
	sr := 48000
	silence := make([]float32, sr*3) // 3 s of it, with a whisper of noise
	rng := rand.New(rand.NewSource(1))
	for i := range silence {
		silence[i] = float32(3e-4 * (2*rng.Float64() - 1))
	}

	cap := &capture{}
	e := NewEngine(defaultCfg(), cap.emit)
	if err := e.Run(context.Background(), &sliceSource{samples: silence, rate: sr}); err != nil {
		t.Fatal(err)
	}

	cap.mu.Lock()
	defer cap.mu.Unlock()
	for _, m := range cap.all {
		if len(m.Reeds) > 0 {
			t.Fatalf("named %s (%.2f Hz) out of silence at level %.5f",
				m.NoteName, m.Reeds[0].Freq, m.InputLevel)
		}
	}
	if len(cap.all) == 0 {
		t.Fatal("no measurements at all: the level meter would freeze")
	}
}

// A note is still measured when it follows silence, so the guard above cannot simply mute everything.
func TestEngineHearsNoteAfterSilence(t *testing.T) {
	sr := 48000
	quiet := make([]float32, sr*2)
	for i := range quiet {
		quiet[i] = float32(3e-4 * math.Sin(float64(i)))
	}
	tone := make([]float32, 0, sr*6)
	tone = append(tone, quiet...)
	for i := 0; i < sr*6; i++ {
		tone = append(tone, float32(0.4*math.Sin(2*math.Pi*442.0*float64(i)/float64(sr))))
	}

	cap := &capture{}
	e := NewEngine(defaultCfg(), cap.emit)
	if err := e.Run(context.Background(), &sliceSource{samples: tone, rate: sr}); err != nil {
		t.Fatal(err)
	}

	m := cap.lastFull()
	if len(m.Reeds) == 0 {
		t.Fatal("the note after the silence was never measured")
	}
	if math.Abs(m.Reeds[0].Freq-442.0) > 0.05 {
		t.Fatalf("measured %.4f Hz, want 442", m.Reeds[0].Freq)
	}
}

// One reed sounding while the tuner is set to three: it used to answer with three evenly-spaced reeds
// (the ribs of a single lobe) and two invented beats. A reed is only a reed if the spectrum comes back
// down on either side of it.
func TestEngineOneReedIsNotThree(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 8 * time.Second,
		Reeds:    []audio.ReedSpec{{Freq: 331.0, Amp: 0.4, Harmonics: []float64{0.3, 0.15}}},
	}
	cfg := defaultCfg()
	cfg.ReedCount = 3
	m := runEngine(t, spec, cfg).lastFull()

	if len(m.Reeds) != 1 {
		t.Fatalf("one reed sounding, %d reported: %+v", len(m.Reeds), m.Reeds)
	}
	if len(m.Beats) != 0 {
		t.Fatalf("one reed cannot beat with itself: %+v", m.Beats)
	}
	if math.Abs(m.Reeds[0].Freq-331.0) > 0.05 {
		t.Fatalf("freq = %v want 331.0", m.Reeds[0].Freq)
	}
}

// A pair too close to pull apart spectrally still has to be measured: that is the pair a technician is tuning toward unison.
func TestEngineMergedPairIsStillMeasured(t *testing.T) {
	for _, beat := range []float64{0.8, 1.0} {
		spec := audio.SynthSpec{
			Duration: 10 * time.Second,
			Reeds: []audio.ReedSpec{
				{Freq: 440.0, Amp: 0.4},
				{Freq: 440.0 + beat, Amp: 0.38},
			},
		}
		cfg := defaultCfg()
		cfg.ReedCount = 2
		m := runEngine(t, spec, cfg).lastFull()

		if len(m.Beats) != 1 {
			t.Fatalf("%.1f Hz pair: beats = %+v", beat, m.Beats)
		}
		if math.Abs(m.Beats[0].Hz-beat) > 0.08 {
			t.Errorf("%.1f Hz pair: measured %.3f Hz", beat, m.Beats[0].Hz)
		}
		if m.ReedsSeparated {
			t.Errorf("%.1f Hz pair sits inside one lobe: its peaks are not two reeds", beat)
		}
	}
}

// A single reed on hand-worked bellows throws real spectral sidebands (28 and 13 percent of the reed
// on the technician's A3 recording). A line that much quieter than its neighbour is the bellows, not a reed.
func TestEngineSidebandsAreNotReeds(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 8 * time.Second,
		Reeds: []audio.ReedSpec{
			{Freq: 219.48, Amp: 0.28 * 0.4}, // lower sideband
			{Freq: 221.19, Amp: 0.40},       // the reed
			{Freq: 223.01, Amp: 0.13 * 0.4}, // upper sideband
		},
	}
	cfg := defaultCfg()
	cfg.ReedCount = 3
	m := runEngine(t, spec, cfg).lastFull()

	if len(m.Reeds) != 1 {
		t.Fatalf("one reed and its sidebands, %d reeds reported: %+v", len(m.Reeds), m.Reeds)
	}
	if math.Abs(m.Reeds[0].Freq-221.19) > 0.05 {
		t.Fatalf("reported %.3f Hz, want the reed at 221.19", m.Reeds[0].Freq)
	}
	if len(m.Beats) != 0 {
		t.Fatalf("a reed does not beat with its own sidebands: %+v", m.Beats)
	}
}
