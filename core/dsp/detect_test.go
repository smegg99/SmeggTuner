package dsp

import (
	"testing"

	"smegg.me/smeggtuner/core/tuning"
)

func resultWith(pairs map[tuning.Note]float64) CoarseResult {
	var r CoarseResult
	for i := range r.NoteEnergy {
		r.NoteEnergy[i] = 0.0001
	}
	for n, e := range pairs {
		r.NoteEnergy[int(n-tuning.MinNote)] = e
	}
	return r
}

func lowFloor() [105]float64 {
	var f [105]float64
	for i := range f {
		f[i] = 0.0005
	}
	return f
}

func TestDetectSimpleTone(t *testing.T) {
	r := resultWith(map[tuning.Note]float64{69: 0.5, 81: 0.2, 88: 0.1})
	n, ok := NewDetector().Detect(r, lowFloor())
	if !ok || n != tuning.NoteA4 {
		t.Fatalf("got %v ok=%v want A4", n, ok)
	}
}

func TestDetectWeakFundamental(t *testing.T) {
	// A1 reed: fundamental barely there, harmonics strong
	r := resultWith(map[tuning.Note]float64{
		33: 0.01, // A1 fundamental, weak but present
		45: 0.5,  // 2nd harmonic (A2)
		52: 0.4,  // 3rd (E3)
		57: 0.3,  // 4th (A3)
	})
	n, ok := NewDetector().Detect(r, lowFloor())
	if !ok || n != tuning.Note(33) {
		t.Fatalf("got %v ok=%v want A1(33)", n, ok)
	}
}

func TestDetectSilence(t *testing.T) {
	r := resultWith(nil)
	if _, ok := NewDetector().Detect(r, lowFloor()); ok {
		t.Fatal("silence must not detect")
	}
}

// A bassoon (16') G#3 reed whose fundamental is buried and whose fifth harmonic is the loudest thing:
// the tuner used to report C6, five harmonics up.
func TestDetectBassoonReedReadAsHarmonic(t *testing.T) {
	const gSharp3 = tuning.Note(56)
	r := resultWith(map[tuning.Note]float64{
		gSharp3:      0.02, // fundamental, present but quiet
		gSharp3 + 12: 0.10, // 2nd
		gSharp3 + 19: 0.18, // 3rd
		gSharp3 + 24: 0.22, // 4th
		gSharp3 + 28: 0.50, // 5th, the loudest partial: C6
		gSharp3 + 31: 0.15, // 6th
	})
	n, ok := NewDetector().Detect(r, lowFloor())
	if !ok || n != gSharp3 {
		t.Fatalf("got %v (ok=%v), want G#3 (%v): the fifth harmonic won", n, ok, gSharp3)
	}
}

// A 16' Bugari reed as the coarse map really sees it: the fundamental 24 dB under the loudest line,
// the 7th partial invisible because it falls 33 cents off the note grid, and the NINTH partial
// winning. One step down the series lands on the 3rd partial (A4) - the answer the tuner used to
// give - so the walk has to keep going. Energies are per-note amplitudes over the loudest, and the
// floor is the leak guard the engine applies.
func TestDetectSixteenFootUnderItsNinthPartial(t *testing.T) {
	const d3 = tuning.Note(50)
	r := resultWith(map[tuning.Note]float64{
		d3:      0.115, // fundamental
		d3 + 12: 0.130, // 2nd
		d3 + 19: 0.400, // 3rd, A4
		d3 + 24: 0.430, // 4th, carrying the 4' rank as well
		d3 + 28: 0.240, // 5th
		d3 + 31: 0.480, // 6th
		d3 + 34: 0.040, // 7th, off the tempered grid and all but gone
		d3 + 36: 0.180, // 8th
		d3 + 38: 1.000, // 9th, the loudest note in the map
	})
	var floor [tuning.NumNotes]float64
	for i := range floor {
		floor[i] = leakGuardRatio
	}
	n, ok := NewDetector().Detect(r, floor)
	if !ok || n != d3 {
		t.Fatalf("got %v (ok=%v), want D3 (%v): the walk stopped on a partial", n, ok, d3)
	}
}

// The correction must not invent a fundamental under a note that has none: a clean high tone stays put.
func TestDetectHighToneKeepsItsPitch(t *testing.T) {
	const c6 = tuning.Note(84)
	r := resultWith(map[tuning.Note]float64{
		c6:      0.50,
		c6 + 12: 0.15,
		c6 + 19: 0.08,
	})
	n, ok := NewDetector().Detect(r, lowFloor())
	if !ok || n != c6 {
		t.Fatalf("got %v (ok=%v), want C6 (%v)", n, ok, c6)
	}
}
