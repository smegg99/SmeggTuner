// Package tests holds cross-package golden tests asserting the end-to-end accuracy contract.
package tests

import (
	"context"
	"math"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/audio"
	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/tuning"
)

func run(t *testing.T, spec audio.SynthSpec, cfg dsp.EngineConfig) dsp.Measurement {
	t.Helper()
	var last dsp.Measurement
	var locked *dsp.Measurement
	eng := dsp.NewEngine(cfg, func(m dsp.Measurement) {
		// The engine emits reedless heartbeats between results; keep the last tick that actually measured.
		if len(m.Reeds) > 0 {
			last = m
		}
		if m.Locked {
			c := m
			locked = &c
		}
	})
	if err := eng.Run(context.Background(), audio.NewSynthSource(spec)); err != nil {
		t.Fatal(err)
	}
	if locked != nil {
		return *locked
	}
	return last
}

func cfg(reeds int) dsp.EngineConfig {
	// CalibSecs 0: synthetic input starts mid-note, calibration stays off
	return dsp.EngineConfig{
		A4: 440, ReedCount: reeds, FineWindow: 3 * time.Second,
	}
}

func TestGoldenSingleReedAcrossRange(t *testing.T) {
	freqs := []float64{55, 110, 220, 440, 880, 1760, 3520}
	for _, f := range freqs {
		m := run(t, audio.SynthSpec{
			Duration: 6 * time.Second,
			Reeds:    []audio.ReedSpec{{Freq: f, Amp: 0.4, Harmonics: []float64{0.3, 0.15, 0.08}}},
		}, cfg(1))
		if len(m.Reeds) != 1 {
			t.Fatalf("f=%v: reeds=%+v", f, m.Reeds)
		}
		if math.Abs(m.Reeds[0].Freq-f) > 0.05 {
			t.Errorf("f=%v: measured %v (err %.4f Hz)", f, m.Reeds[0].Freq, m.Reeds[0].Freq-f)
		}
		t.Logf("f=%v: note=%s measured=%.5f err=%+.5f Hz locked=%v",
			f, m.NoteName, m.Reeds[0].Freq, m.Reeds[0].Freq-f, m.Locked)
	}
}

func TestGoldenLowE1(t *testing.T) {
	f := tuning.Note(28).Freq(440) // E1 41.2 Hz
	m := run(t, audio.SynthSpec{
		Duration: 8 * time.Second,
		Reeds:    []audio.ReedSpec{{Freq: f, Amp: 0.05, Harmonics: []float64{1.5, 1.2, 0.8}}},
	}, cfg(1))
	if m.Note != tuning.Note(28) {
		t.Fatalf("weak-fundamental low reed detected as %s", m.NoteName)
	}
	if len(m.Reeds) == 0 {
		t.Fatalf("no reeds measured")
	}
	if math.Abs(m.Reeds[0].Freq-f) > 0.1 {
		t.Fatalf("E1 freq err %.4f", m.Reeds[0].Freq-f)
	}
	t.Logf("E1: note=%s target=%.5f measured=%.5f err=%+.5f Hz locked=%v",
		m.NoteName, f, m.Reeds[0].Freq, m.Reeds[0].Freq-f, m.Locked)
}

func TestGoldenThreeReeds(t *testing.T) {
	m := run(t, audio.SynthSpec{
		Duration: 6 * time.Second,
		Reeds: []audio.ReedSpec{
			{Freq: 437.4, Amp: 0.35},
			{Freq: 440.0, Amp: 0.4},
			{Freq: 442.6, Amp: 0.35},
		},
	}, cfg(3))
	if len(m.Reeds) != 3 || len(m.Beats) != 2 {
		t.Fatalf("reeds=%d beats=%d", len(m.Reeds), len(m.Beats))
	}
	for i, want := range []float64{437.4, 440.0, 442.6} {
		if math.Abs(m.Reeds[i].Freq-want) > 0.05 {
			t.Errorf("reed %d: %v want %v", i+1, m.Reeds[i].Freq, want)
		}
		t.Logf("reed %d: measured=%.5f want=%.1f err=%+.5f Hz",
			i+1, m.Reeds[i].Freq, want, m.Reeds[i].Freq-want)
	}
	for i, want := range []float64{2.6, 2.6} {
		if math.Abs(m.Beats[i].Hz-want) > 0.05 {
			t.Errorf("beat %s: %v want %v", m.Beats[i].Pair, m.Beats[i].Hz, want)
		}
		t.Logf("beat %s: measured=%.5f want=%.1f err=%+.5f Hz fromEnvelope=%v",
			m.Beats[i].Pair, m.Beats[i].Hz, want, m.Beats[i].Hz-want, m.Beats[i].FromEnvelope)
	}
}

func TestGoldenNoiseRobust(t *testing.T) {
	m := run(t, audio.SynthSpec{
		Duration: 6 * time.Second,
		Reeds:    []audio.ReedSpec{{Freq: 440, Amp: 0.4, Harmonics: []float64{0.3}}},
		NoiseAmp: 0.04, // ~20 dB SNR
		Seed:     42,
	}, cfg(1))
	if len(m.Reeds) != 1 || math.Abs(m.Reeds[0].Freq-440) > 0.1 {
		t.Fatalf("noisy measurement: %+v", m.Reeds)
	}
	t.Logf("noise: note=%s measured=%.5f err=%+.5f Hz locked=%v",
		m.NoteName, m.Reeds[0].Freq, m.Reeds[0].Freq-440, m.Locked)
}

func TestGoldenHumRejection(t *testing.T) {
	c := cfg(1)
	c.Hum50 = true
	m := run(t, audio.SynthSpec{
		Duration: 6 * time.Second,
		Reeds:    []audio.ReedSpec{{Freq: 440, Amp: 0.3}},
		HumFreq:  50,
		HumAmp:   0.2,
	}, c)
	if m.Note != tuning.NoteA4 {
		t.Fatalf("hum stole detection: %s", m.NoteName)
	}
	if len(m.Reeds) == 0 {
		t.Fatalf("no reeds measured")
	}
	if math.Abs(m.Reeds[0].Freq-440) > 0.05 {
		t.Fatalf("hum shifted measurement: %v", m.Reeds[0].Freq)
	}
	t.Logf("hum: note=%s measured=%.5f err=%+.5f Hz locked=%v",
		m.NoteName, m.Reeds[0].Freq, m.Reeds[0].Freq-440, m.Locked)
}
