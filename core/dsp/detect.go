package dsp

import "smegg.me/smeggtuner/core/tuning"

// Harmonics 1..7 in equal-tempered semitone steps and their score weights.
var harmonicOffsets = [7]int{0, 12, 19, 24, 28, 31, 34}
var harmonicWeights = [7]float64{1.0, 0.5, 0.33, 0.25, 0.2, 0.17, 0.14}

// Where a fundamental sits if the picked note is really its 6th..2nd harmonic. Deepest first, so a
// weak fundamental wins over the harmonic that drowned it out.
var subHarmonicOffsets = [5]int{31, 28, 24, 19, 12}

// subHarmonicSupport is the share of the winner's score a candidate fundamental must still carry to
// be taken for the reed. A 16' rank whose fundamental sits 24 dB under its loudest partial scores
// only about half of the partial that drowned it: the Bugari D3 take measures 0.52 through the
// middle of its run and 0.44 at its worst, so a bar at 0.5 throws away half of it. Dropping further,
// to 0.40, starts costing the older recordings notes they used to get right.
const subHarmonicSupport = 0.45

// Detector picks the sounding note from coarse energies. It is stateless; hold hysteresis lives in the engine.
type Detector struct{}

func NewDetector() *Detector { return &Detector{} }

func (d *Detector) Detect(e CoarseResult, floor [tuning.NumNotes]float64) (tuning.Note, bool) {
	snr := func(i int) float64 {
		if i < 0 || i >= tuning.NumNotes {
			return 0
		}
		return e.NoteEnergy[i] / floor[i]
	}
	score := func(i int) float64 {
		var s float64
		for k, off := range harmonicOffsets {
			s += harmonicWeights[k] * snr(i+off)
		}
		return s
	}
	best, bestScore := -1, 0.0
	for i := 0; i < tuning.NumNotes; i++ {
		if snr(i) < 6 {
			continue
		}
		if s := score(i); s > bestScore {
			best, bestScore = i, s
		}
	}
	if best < 0 {
		return 0, false
	}
	// A bassoon reed puts little energy in its fundamental and much in the partials above, so the
	// winner is often a harmonic of the note actually sounding. Walk back down the harmonic series,
	// deepest first, taking the lowest note that is both audible and carries a fundamental's support.
	//
	// It repeats because one step undoes one factor: from a 9th partial a single step lands on the
	// 3rd, still not the reed. It always ends - every step drops at least an octave, and bestScore
	// stays the original winner's, so each step down has more to prove than the last.
	for stepped := true; stepped; {
		stepped = false
		for _, off := range subHarmonicOffsets {
			cand := best - off
			if cand < 0 {
				continue
			}
			if snr(cand) > 2 && score(cand) >= subHarmonicSupport*bestScore {
				best, stepped = cand, true
				break
			}
		}
	}
	return tuning.MinNote + tuning.Note(best), true
}
