package dsp

import (
	"context"
	"math"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/audio"
	"smegg.me/smeggtuner/core/tuning"
)

// The register a musette lives in and the one that is unmeasurable without this: a tremolo held at a
// constant number of cents shrinks in hertz as the notes fall, so a pair 2.3 Hz apart at C4 is 1.1 Hz
// apart at C3, inside the window's main lobe (4/T, 1.33 Hz).
const (
	noteC3 = tuning.Note(48)
	noteC2 = tuning.Note(36)
)

// pairSpec is two reeds beating, with harmonics because a pair of pure tones at C2 has so little
// energy in its fundamental that note detection flickers a semitone at the beat's null. The zoom band
// is +-16 Hz, so nothing above the fundamental reaches the stage under test.
func pairSpec(centre, beat, ratio float64, secs int) audio.SynthSpec {
	h := []float64{0.3, 0.15}
	return audio.SynthSpec{
		Duration: time.Duration(secs) * time.Second,
		Reeds: []audio.ReedSpec{
			{Freq: centre - beat/2, Amp: 0.25, Harmonics: h},
			{Freq: centre + beat/2, Amp: 0.25 * ratio, Harmonics: h},
		},
	}
}

// reedsWithin checks both reed frequencies against the truth, in cents.
func reedsWithin(t *testing.T, m Measurement, lo, hi, tol float64) {
	t.Helper()
	if len(m.Reeds) != 2 {
		t.Fatalf("expected two reeds, got %d: %+v", len(m.Reeds), m.Reeds)
	}
	for i, want := range []float64{lo, hi} {
		got := m.Reeds[i].Freq
		if c := tuning.Cents(got, want); math.Abs(c) > tol {
			t.Errorf("reed %d: %.4f Hz, want %.4f (%.2f cents out, tolerance %.1f)",
				i+1, got, want, c, tol)
		}
	}
}

// A merged pair is recovered from its beat, to within a cent, all the way down the register where the
// spectrum has given up. The 2 Hz pairs still separate spectrally at both notes: the same three beats
// have to come out right whichever route the engine takes.
func TestEngineMergedPairRecoveredFromBeat(t *testing.T) {
	for _, note := range []tuning.Note{noteC3, noteC2} {
		for _, beat := range []float64{0.8, 1.2, 2.0} {
			centre := note.Freq(440)
			cfg := defaultCfg()
			cfg.ReedCount = 2
			m := runEngine(t, pairSpec(centre, beat, 1.0, 10), cfg).lastFull()

			if m.Note != note {
				t.Fatalf("%s beat %.1f: tracked %s", note.Name(tuning.NamingCDEFGAB), beat, m.NoteName)
			}
			if !m.ReedsSeparated && !m.ReedsFromBeat {
				t.Fatalf("%s beat %.1f: no per-reed answer at all: %+v", m.NoteName, beat, m)
			}
			if m.ReedsSeparated && m.ReedsFromBeat {
				t.Fatalf("%s beat %.1f: the two flags are two different routes and cannot both hold", m.NoteName, beat)
			}
			reedsWithin(t, m, centre-beat/2, centre+beat/2, 1.0)
			t.Logf("%-3s beat %.1f Hz: fromBeat=%v separated=%v reeds %.4f / %.4f",
				m.NoteName, beat, m.ReedsFromBeat, m.ReedsSeparated, m.Reeds[0].Freq, m.Reeds[1].Freq)
		}
	}
}

// The pair the equal-amplitude assumption gets wrong: the composite peak is pulled towards the louder
// reed, so "the peak, plus and minus half a beat" is out by up to half a beat, and out in a different
// direction depending on which reed is loud. Both orientations are here for that reason.
func TestEngineMergedPairWithUnequalReeds(t *testing.T) {
	centre := noteC3.Freq(440)
	const beat = 1.2
	lo, hi := centre-beat/2, centre+beat/2

	for _, tc := range []struct {
		name  string
		ratio float64
	}{
		{"lower reed twice the upper", 0.5},
		{"upper reed twice the lower", 2.0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := defaultCfg()
			cfg.ReedCount = 2
			m := runEngine(t, pairSpec(centre, beat, tc.ratio, 10), cfg).lastFull()

			if !m.ReedsFromBeat {
				t.Fatalf("a 2:1 pair is still a pair: %+v", m)
			}
			reedsWithin(t, m, lo, hi, 1.0)

			// Teeth: the naive reconstruction (peak +- half a beat) lands nowhere near either reed.
			peak := mergedPeak(t, pairSpec(centre, beat, tc.ratio, 6), centre)
			naiveLo := tuning.Cents(peak-beat/2, lo)
			naiveHi := tuning.Cents(peak+beat/2, hi)
			if math.Abs(naiveLo) < 1.0 && math.Abs(naiveHi) < 1.0 {
				t.Fatalf("peak %.4f Hz +- half a beat already lands within a cent of both reeds "+
					"(%.2f / %.2f cents), so this test would pass without doing anything",
					peak, naiveLo, naiveHi)
			}
			t.Logf("composite peak %.4f Hz: naive peak+-b/2 would be %+.2f / %+.2f cents out; "+
				"recovered %.4f / %.4f Hz (%+.2f / %+.2f cents)",
				peak, naiveLo, naiveHi, m.Reeds[0].Freq, m.Reeds[1].Freq,
				tuning.Cents(m.Reeds[0].Freq, lo), tuning.Cents(m.Reeds[1].Freq, hi))
		})
	}
}

// mergedPeak is the single line the peak picker finds where the pair is.
func mergedPeak(t *testing.T, spec audio.SynthSpec, centre float64) float64 {
	t.Helper()
	src := audio.NewSynthSource(spec)
	blocks, err := src.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	sr := src.Info().SampleRate
	ring := NewRing(sr * 10)
	for b := range blocks {
		ring.Write(b.Samples)
	}
	zr := NewZoom(sr).Analyze(ring, centre, math.Max(16, centre*0.035), 3*time.Second)
	peaks := FindPeaks(zr, 2, 4.0/3.0)
	if len(peaks) != 1 {
		t.Fatalf("a merged pair is one line to the spectrum, found %d", len(peaks))
	}
	return peaks[0].Freq
}
