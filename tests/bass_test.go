package tests

// Two five-voice bass buttons at A440; the rungs are independent FFT measurements, every other line a harmonic of one.

import (
	"fmt"
	"math"
	"testing"

	"smegg.me/smeggtuner/backend/core/tuning"
)

const (
	bassA4 = 440.0
	// Hand-tuned rungs are near octaves; the widest gap measured is 11.1 cents.
	bassOctaveTolCents = 30.0
	// Frequencies drift with the bellows across a run; 1.5 Hz at the top rung is under 2 cents.
	bassFreqTolHz = 1.5
)

type bassTake struct {
	file   string
	voices int
	key    tuning.Note // the lowest rung, which names the note
	rungs  []float64
	stray  string // why the reference scan misreads it in file playback; see strayLine
}

var bassTakes = []bassTake{{
	file: "PrimeBass C.wav", voices: 5, key: 36, rungs: []float64{66.016, 131.697, 262.038, 526.263, 1048.510},
}, {
	file: "PrimaBass F.wav", voices: 5, key: 41, rungs: []float64{87.904, 174.687, 350.163, 699.555, 1398.954},
	stray: "a 43.9 Hz line (F1, half the lowest rung) scores 30-42; the lowest rung F2 scores 70-122, under 30x it, the third rung F4 895-1367 clears it",
}}

func TestTheBassLadderIsHeardRungByRung(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("loads recordings")
	}
	for _, take := range bassTakes {
		t.Run(take.file, func(t *testing.T) {
			r := runFile(t, "../sounds/"+take.file, bassConfig(t, bassA4, take.voices))
			key := take.key
			if !reads(r.all, key) && take.stray != "" {
				t.Skipf("the button is %s: %s (reads %s)", name(key), take.stray, summary(r.all))
			}
			if !reads(r.all, key) {
				t.Fatalf("reads %s, want %s", summary(r.all), name(key))
			}
			if take.stray != "" {
				t.Errorf("listed as a stray-line misreading but reads %s: drop its stray note", name(key))
			}
			frames := sounding(r.all, key)
			got := make([]float64, len(take.rungs))
			for i, want := range take.rungs {
				for _, m := range frames {
					if m.Reeds[i].Present && m.Reeds[i].Octave != 12*i {
						t.Fatalf("rung %d at octave %d, want %d", i+1, m.Reeds[i].Octave, 12*i)
					}
				}
				var n int
				got[i], n = sustained(frames, i)
				if n == 0 || math.Abs(got[i]-want) > bassFreqTolHz {
					t.Errorf("rung %d at %.3f Hz over %d of %d analyses, want %.3f +-%.1f", i+1, got[i], n, len(frames), want, bassFreqTolHz)
				}
				t.Logf("rung %d: %.3f Hz over %d of %d analyses", i+1, got[i], n, len(frames))
			}
			// A rung read off the one below's second partial would sit at exactly 0 cents, a missed one 1200 out.
			for i := 1; i < len(got); i++ {
				if c := cents(got[i], 2*got[i-1]); math.Abs(c) > bassOctaveTolCents {
					t.Errorf("rungs %d and %d are %+.1f cents off an octave", i, i+1, c)
				}
			}
			if ts := takes(r.all); len(ts) == 0 {
				t.Logf("never locked over %d analyses", len(frames))
			}
		})
	}
}

// Beats are reported between neighbouring rungs only, named by their slots, and only where both sounded.
func TestTheBassLadderOnlyBeatsBetweenNeighbouringRungs(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("loads recordings")
	}
	for _, take := range bassTakes {
		t.Run(take.file, func(t *testing.T) {
			for _, m := range measured(runFile(t, "../sounds/"+take.file, bassConfig(t, bassA4, take.voices)).all) {
				if len(m.Beats) != take.voices-1 {
					t.Fatalf("%d beats for %d rungs", len(m.Beats), take.voices)
				}
				for i, b := range m.Beats {
					if b.Pair != fmt.Sprintf("%d-%d", i+1, i+2) {
						t.Fatalf("beat %d names pair %q", i, b.Pair)
					}
					if b.Present != (m.Reeds[i].Present && m.Reeds[i+1].Present) || b.Present && (math.IsNaN(b.Hz) || b.Hz < 0) {
						t.Fatalf("beat %s %+v over rungs %+v and %+v", b.Pair, b, m.Reeds[i], m.Reeds[i+1])
					}
				}
			}
		})
	}
}
