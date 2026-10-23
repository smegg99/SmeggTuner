package tests

import (
	"math"
	"testing"
	"time"

	"smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/tuning"
)

func runSynth(t *testing.T, spec audio.SynthSpec, cfg dsp.EngineConfig) []dsp.Measurement {
	t.Helper()
	return runSource(t, cfg, audio.NewSynthSource(spec)).all
}

// firstTake is the reading a held note files; a steady synthetic note must lock.
func firstTake(t *testing.T, all []dsp.Measurement) dsp.Measurement {
	t.Helper()
	ts := takes(all)
	if len(ts) == 0 {
		t.Fatalf("a steady note never locked: %s", summary(all))
	}
	return ts[0]
}

func TestGoldenSingleReedAcrossRange(t *testing.T) {
	t.Parallel()
	for _, want := range []tuning.Note{33, 45, 57, 69, 81, 93, 105} { // A1..A7
		f := want.Freq(440)
		m := firstTake(t, runSynth(t, audio.SynthSpec{Duration: 8 * time.Second,
			Reeds: []audio.ReedSpec{{Freq: f, Amp: 0.4, Harmonics: []float64{0.3, 0.15, 0.08}}}}, dsp.EngineConfig{A4: 440}))
		if m.Note != want || !m.Reeds[0].Present || math.Abs(m.Reeds[0].Freq-f) > 0.05 {
			t.Errorf("%.2f Hz: took %s %+v", f, name(m.Note), m.Reeds[0])
		}
		t.Logf("%s: %.5f Hz (%+.5f) %+.3f units", name(m.Note), m.Reeds[0].Freq, m.Reeds[0].Freq-f, m.Reeds[0].Dev)
	}
}

// Partials outweighing the fundamental do not move the scan up; only a note scoring 30x the lowest does.
func TestGoldenLowE1(t *testing.T) {
	t.Parallel()
	want := tuning.Note(28)
	f := want.Freq(440)
	m := firstTake(t, runSynth(t, audio.SynthSpec{Duration: 8 * time.Second,
		Reeds: []audio.ReedSpec{{Freq: f, Amp: 0.05, Harmonics: []float64{1.5, 1.2, 0.8}}}}, dsp.EngineConfig{A4: 440}))
	if m.Note != want || math.Abs(m.Reeds[0].Freq-f) > 0.1 {
		t.Fatalf("E1 took as %s %+v", name(m.Note), m.Reeds[0])
	}
	t.Logf("E1: %.5f Hz (%+.5f)", m.Reeds[0].Freq, m.Reeds[0].Freq-f)
}

func TestGoldenThreeReeds(t *testing.T) {
	t.Parallel()
	freqs := []float64{437.4, 440.0, 442.6}
	m := firstTake(t, runSynth(t, audio.SynthSpec{Duration: 8 * time.Second, Reeds: []audio.ReedSpec{
		{Freq: freqs[0], Amp: 0.35}, {Freq: freqs[1], Amp: 0.4}, {Freq: freqs[2], Amp: 0.35}}}, dsp.EngineConfig{A4: 440, ReedCount: 3}))
	if m.Note != tuning.NoteA4 || len(m.Reeds) != 3 || len(m.Beats) != 2 {
		t.Fatalf("took %s with %d slots and %d beats", name(m.Note), len(m.Reeds), len(m.Beats))
	}
	for i, want := range freqs {
		if r := m.Reeds[i]; !r.Present || math.Abs(r.Freq-want) > 0.05 {
			t.Errorf("slot %d: %+v want %.1f", i+1, r, want)
		}
		t.Logf("slot %d: %.5f Hz (%+.5f)", i+1, m.Reeds[i].Freq, m.Reeds[i].Freq-want)
	}
	for i, b := range m.Beats {
		if want := freqs[i+1] - freqs[i]; !b.Present || math.Abs(b.Hz-want) > 0.05 {
			t.Errorf("beat %s: %+v want %.1f", b.Pair, b, want)
		}
		t.Logf("beat %s: %.5f Hz", b.Pair, b.Hz)
	}
}

func TestGoldenNoiseRobust(t *testing.T) {
	t.Parallel()
	m := firstTake(t, runSynth(t, audio.SynthSpec{Duration: 8 * time.Second, NoiseAmp: 0.04, Seed: 42, // about 20 dB SNR
		Reeds: []audio.ReedSpec{{Freq: 440, Amp: 0.4, Harmonics: []float64{0.3}}}}, dsp.EngineConfig{A4: 440}))
	if m.Note != tuning.NoteA4 || math.Abs(m.Reeds[0].Freq-440) > 0.1 {
		t.Fatalf("took %s %+v", name(m.Note), m.Reeds[0])
	}
	t.Logf("noise: %.5f Hz (%+.5f)", m.Reeds[0].Freq, m.Reeds[0].Freq-440)
}

// Mains hum is a strong steady line under every note; the hum filter is what keeps the scan off it.
func TestGoldenHumRejection(t *testing.T) {
	t.Parallel()
	spec := audio.SynthSpec{Duration: 8 * time.Second, HumFreq: 50, HumAmp: 0.2, Reeds: []audio.ReedSpec{{Freq: 440, Amp: 0.3}}}
	m := firstTake(t, runSynth(t, spec, dsp.EngineConfig{A4: 440, Hum50: true}))
	if m.Note != tuning.NoteA4 || math.Abs(m.Reeds[0].Freq-440) > 0.05 {
		t.Fatalf("with the 50 Hz filter took %s %+v", name(m.Note), m.Reeds[0])
	}
	t.Logf("hum filtered: %.5f Hz (%+.5f)", m.Reeds[0].Freq, m.Reeds[0].Freq-440)
	if all := runSynth(t, spec, dsp.EngineConfig{A4: 440}); !reads(all, 31) {
		t.Errorf("without the filter the 50 Hz line (G1) holds the scan, and it reads %s", summary(all))
	}
}
