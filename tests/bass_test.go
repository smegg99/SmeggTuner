package tests

// The bass side on real audio. sounds/PrimeBass C.wav and sounds/PrimaBass F.wav are the only two
// recordings of a bass machine in the repository and nothing tested them - the bass path had unit
// tests over synthesised ladders and none over a button anybody actually pressed.
//
// A bass button sounds ONE pitch class on every rank of the machine at once, an octave apart per
// rank. That is what makes it hard: the low rungs put little energy in their fundamental and a lot
// in the partials above, so a 32' C2 shows a strong 3rd partial at G3 and a strong 5th at E4. Read
// naively that is a C major chord, and a detector that takes the loudest line for the reed reports
// the wrong pitch class entirely. Every line in both recordings was checked against an independent
// FFT: each is a harmonic of one of the C (or F) rungs, and no other pitch class is sounding.

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	"smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/tuning"
)

type bassTake struct {
	file string
	// voices the machine sounds; both recordings carry five, 32' through 2'.
	voices int
	// the lowest rung, which is the note the engine reports.
	key tuning.Note
	// what each rung measured, lowest first, through an independent FFT and again through the
	// engine - the two agreed within 0.1 Hz on every rung.
	rungs []float64
}

// Measured at A440: the machine is a shade sharp of it, which is the instrument, not a reading error.
const bassA4 = 440.0

// A rung is a reed in its own right, tuned by hand, so consecutive rungs are near an octave rather
// than exactly one. The widest gap measured across both recordings is 11.1 cents.
const bassOctaveTolCents = 30.0

// Frequencies drift with the bellows across a run; 1 Hz at the top rung is under 1.5 cents.
const bassFreqTolHz = 1.5

var bassTakes = []bassTake{{
	file:   "PrimeBass C.wav",
	voices: 5,
	key:    tuning.Note(36), // C2
	rungs:  []float64{66.016, 131.697, 262.038, 526.263, 1048.510},
}, {
	file:   "PrimaBass F.wav",
	voices: 5,
	key:    tuning.Note(41), // F2
	rungs:  []float64{87.904, 174.687, 350.163, 699.555, 1398.954},
}}

func TestTheBassLadderIsHeardRungByRung(t *testing.T) {
	if testing.Short() {
		t.Skip("loads multi-megabyte recordings")
	}

	for _, take := range bassTakes {
		t.Run(take.file, func(t *testing.T) {
			feet := session.BassFeet(take.voices)
			if feet == nil {
				t.Fatalf("%d is not a bass machine this app describes", take.voices)
			}
			m, ok := measureUnderLayout(t, "../sounds/"+take.file, bassA4,
				len(feet), session.OctavesOfFeet(feet))
			if !ok {
				t.Fatalf("no usable reading")
			}

			// The pitch class is the whole game here: the strong 3rd and 5th partials of the low
			// rungs spell a major chord, and naming the note off one of them is the failure.
			if m.Note != take.key {
				t.Fatalf("read %s, want %s - the ladder is one pitch class, and the fifth and third "+
					"above it are the low rungs' partials, not notes",
					m.Note.Name(tuning.NamingCDEFGAB), take.key.Name(tuning.NamingCDEFGAB))
			}

			if len(m.Reeds) != len(take.rungs) {
				t.Fatalf("resolved %d rungs, the machine sounds %d: %+v",
					len(m.Reeds), len(take.rungs), m.Reeds)
			}

			for i, want := range take.rungs {
				r := m.Reeds[i]
				if octave := 12 * i; r.Octave != octave {
					t.Errorf("rung %d (%s) sits at octave %d, want %d",
						i+1, feetLabel(feet, i), r.Octave, octave)
				}
				if math.Abs(r.Freq-want) > bassFreqTolHz {
					t.Errorf("rung %d (%s) measured %.3f Hz, want %.3f +-%.1f",
						i+1, feetLabel(feet, i), r.Freq, want, bassFreqTolHz)
				}
			}

			// Each rung is its own reed, so the ladder is only near-octaves - but a rung read off
			// the rung below's second partial would sit at exactly 0.0 cents, and a rung missed
			// altogether would be 1200 out.
			for i := 1; i < len(m.Reeds); i++ {
				cents := 1200 * math.Log2(m.Reeds[i].Freq/m.Reeds[i-1].Freq/2)
				if math.Abs(cents) > bassOctaveTolCents {
					t.Errorf("rungs %d and %d are %+.1f cents off an octave (%.2f Hz over %.2f Hz)",
						i, i+1, cents, m.Reeds[i].Freq, m.Reeds[i-1].Freq)
				}
			}

			// Every rung of the machine is sounding, so no band may report itself as nothing but
			// the harmonic of the rank beneath it.
			if len(m.Bands) != len(feet) {
				t.Fatalf("want a band per rank, got %+v", m.Bands)
			}
			for i, b := range m.Bands {
				if b.Found != 1 || b.GhostOnly {
					t.Errorf("rung %d (%s) reported %+v, want one confirmed voice",
						i+1, feetLabel(feet, i), b)
				}
			}
		})
	}
}

// A bass button sounds every rank at once, so a beat between two rungs is an octave beat, not a
// musette one. It must never be reported between ranks that are not neighbours.
func TestTheBassLadderOnlyBeatsBetweenNeighbouringRungs(t *testing.T) {
	if testing.Short() {
		t.Skip("loads multi-megabyte recordings")
	}

	for _, take := range bassTakes {
		t.Run(take.file, func(t *testing.T) {
			feet := session.BassFeet(take.voices)
			m, ok := measureUnderLayout(t, "../sounds/"+take.file, bassA4,
				len(feet), session.OctavesOfFeet(feet))
			if !ok {
				t.Fatalf("no usable reading")
			}
			if len(m.Beats) > len(m.Reeds)-1 {
				t.Fatalf("%d beats for %d rungs is more pairs than there are neighbours",
					len(m.Beats), len(m.Reeds))
			}
			for _, b := range m.Beats {
				var lo, hi int
				if _, err := fmt.Sscanf(b.Pair, "%d-%d", &lo, &hi); err != nil {
					t.Errorf("beat pair %q is not two reed numbers: %v", b.Pair, err)
					continue
				}
				if hi != lo+1 {
					t.Errorf("beat %s spans rungs %d and %d, which are not neighbours", b.Pair, lo, hi)
				}
				if math.IsNaN(b.Hz) || math.IsInf(b.Hz, 0) {
					t.Errorf("beat %s is %v", b.Pair, b.Hz)
				}
			}
		})
	}
}

// feetLabel names a rung by its foot, for failures that have to be read at a bench.
func feetLabel(feet []int, i int) string {
	if i < 0 || i >= len(feet) {
		return "?"
	}
	return strconv.Itoa(feet[i]) + "'"
}
