package dsp

import (
	"context"
	"math"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/audio"
)

// synthRing runs a synthetic tone to completion and keeps its newest `hold` seconds in a ring.
func synthRing(t *testing.T, spec audio.SynthSpec, hold time.Duration) *Ring {
	t.Helper()
	if spec.SampleRate == 0 {
		spec.SampleRate = 48000
	}
	src := audio.NewSynthSource(spec)
	ch, err := src.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	ring := NewRing(int(hold.Seconds()*float64(spec.SampleRate)) + spec.SampleRate)
	for b := range ch {
		ring.Write(b.Samples)
	}
	return ring
}

// A compound register sounds several feet at once; the single fine band cannot see a foot an octave
// away. Here A4 on a 16+8+4: the 8' sounds 440, the 16' 220, the 4' 880, all exact octaves.
func TestAnalyzeOctavesRecoversACompoundRegister(t *testing.T) {
	const sr = 48000
	spec := audio.SynthSpec{
		SampleRate: sr,
		Duration:   4 * time.Second,
		Reeds: []audio.ReedSpec{
			{Freq: 220, Amp: 0.40, Harmonics: []float64{0.30, 0.15}}, // 16', with 2nd (440) and 3rd (660)
			{Freq: 440, Amp: 0.50, Harmonics: []float64{0.25}},       // 8', with 2nd (880)
			{Freq: 880, Amp: 0.35},                                   // 4'
		},
		NoiseAmp: 1e-4,
	}
	ring := synthRing(t, spec, 3*time.Second)

	z := NewZoom(sr)
	reqs := []OctaveRequest{{Offset: -12, Reeds: 1}, {Offset: 0, Reeds: 1}, {Offset: 12, Reeds: 1}}
	bands := AnalyzeOctaves(z, ring, 440, reqs, 1.0, 3*time.Second)

	if len(bands) != 3 {
		t.Fatalf("want 3 bands, got %d", len(bands))
	}

	want := map[int]float64{-12: 220, 0: 440, 12: 880}
	for _, b := range bands {
		if !b.Valid {
			t.Errorf("offset %+d: band not valid", b.Offset)
			continue
		}
		if len(b.Reeds) != 1 {
			t.Errorf("offset %+d: want 1 reed, got %d (%v)", b.Offset, len(b.Reeds), b.Reeds)
			continue
		}
		got := b.Reeds[0].Freq
		if math.Abs(got-want[b.Offset]) > 0.3 {
			t.Errorf("offset %+d: want %.1f Hz, got %.4f Hz", b.Offset, want[b.Offset], got)
		}
	}
}

// The regression: a single band placed only on the played note finds the 8' and nothing of the 16' below or 4' above.
func TestASingleBandMissesTheOtherOctaves(t *testing.T) {
	const sr = 48000
	spec := audio.SynthSpec{
		SampleRate: sr,
		Duration:   4 * time.Second,
		Reeds: []audio.ReedSpec{
			{Freq: 220, Amp: 0.40},
			{Freq: 440, Amp: 0.50},
			{Freq: 880, Amp: 0.35},
		},
		NoiseAmp: 1e-4,
	}
	ring := synthRing(t, spec, 3*time.Second)

	z := NewZoom(sr)
	bands := AnalyzeOctaves(z, ring, 440, []OctaveRequest{{Offset: 0, Reeds: 3}}, 1.0, 3*time.Second)
	if len(bands) != 1 || !bands[0].Valid {
		t.Fatalf("want 1 valid band, got %+v", bands)
	}
	for _, r := range bands[0].Reeds {
		if math.Abs(r.Freq-440) > 16 {
			t.Errorf("single band saw a line at %.2f Hz, outside its own octave", r.Freq)
		}
	}
}

// SubtractHarmonics, resolved case: a 16' at 219 leaks its second partial to 438, a genuine 8' sits at
// 443; the 8' band must keep the 443 and drop the 438.
func TestSubtractHarmonicsDropsAResolvedGhost(t *testing.T) {
	bands := []OctaveBand{
		{Offset: -12, Center: 220, Valid: true, Reeds: []Peak{{Freq: 219, Amp: 0.6}}},
		{Offset: 0, Center: 440, Valid: true, Reeds: []Peak{{Freq: 438, Amp: 0.36}, {Freq: 443, Amp: 0.30}}},
	}
	tol := func(float64) float64 { return 3.0 }
	out := SubtractHarmonics(bands, []int{1, 1}, tol)

	if len(out[0].Reeds) != 1 || math.Abs(out[0].Reeds[0].Freq-219) > 0.01 {
		t.Errorf("16' band: want [219], got %+v", out[0].Reeds)
	}
	if len(out[1].Reeds) != 1 || math.Abs(out[1].Reeds[0].Freq-443) > 0.01 {
		t.Errorf("8' band: want [443] (the ghost at 438 subtracted), got %+v", out[1].Reeds)
	}
}

// The coincident case one spectrum cannot resolve: the 16's partial falls exactly on the genuine 8', so the reed is kept.
func TestSubtractHarmonicsKeepsACoincidentReed(t *testing.T) {
	bands := []OctaveBand{
		{Offset: -12, Center: 220, Valid: true, Reeds: []Peak{{Freq: 220, Amp: 0.6}}},
		{Offset: 0, Center: 440, Valid: true, Reeds: []Peak{{Freq: 440, Amp: 0.4}}},
	}
	tol := func(float64) float64 { return 3.0 }
	out := SubtractHarmonics(bands, []int{1, 1}, tol)

	if len(out[1].Reeds) != 1 || math.Abs(out[1].Reeds[0].Freq-440) > 0.01 {
		t.Errorf("8' band: a coincident reed must survive, want [440], got %+v", out[1].Reeds)
	}
}

// End to end where subtraction earns its keep: the 16's second partial (438) is LOUDER than the
// genuine, detuned 8' (443). A single band reports the ghost; AnalyzeCompound reports the real 8'.
func TestAnalyzeCompoundUnmasksAReedUnderALouderGhost(t *testing.T) {
	const sr = 48000
	spec := audio.SynthSpec{
		SampleRate: sr,
		Duration:   4 * time.Second,
		Reeds: []audio.ReedSpec{
			{Freq: 219, Amp: 0.60, Harmonics: []float64{0.60}}, // 16', 2nd partial at 438 (amp 0.36)
			{Freq: 443, Amp: 0.30},                             // genuine 8', detuned, quieter than the ghost
		},
		NoiseAmp: 1e-4,
	}
	ring := synthRing(t, spec, 3*time.Second)
	z := NewZoom(sr)
	reqs := []OctaveRequest{{Offset: -12, Reeds: 1}, {Offset: 0, Reeds: 1}}

	naive := AnalyzeOctaves(z, ring, 440, []OctaveRequest{{Offset: 0, Reeds: 1}}, 1.0, 3*time.Second)
	if len(naive[0].Reeds) != 1 || math.Abs(naive[0].Reeds[0].Freq-438) > 0.5 {
		t.Fatalf("precondition: the naive band should latch the 438 ghost, got %+v", naive[0].Reeds)
	}

	got := z.AnalyzeCompound(ring, 440, reqs, 1.0, 3*time.Second, CentsWindow(15))
	if len(got) != 2 {
		t.Fatalf("want 2 bands, got %d", len(got))
	}
	if len(got[1].Reeds) != 1 || math.Abs(got[1].Reeds[0].Freq-443) > 0.5 {
		t.Errorf("8' band: want the real reed at 443, ghost subtracted, got %+v", got[1].Reeds)
	}
	if len(got[0].Reeds) != 1 || math.Abs(got[0].Reeds[0].Freq-219) > 0.5 {
		t.Errorf("16' band: want [219], got %+v", got[0].Reeds)
	}
}
