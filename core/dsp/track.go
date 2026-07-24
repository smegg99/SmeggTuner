package dsp

import (
	"smegg.me/smeggtuner/core/tuning"
)

// How far back the phase reference sits, in seconds, and how far it may drift before it is too old to
// be the same reed. RefinePhase divides a phase advance by the gap it accumulated over, so the gap
// buys the precision; a quarter of a second is what the fine stage used to be paced at.
const (
	refineHop    = 0.25
	maxRefineHop = 0.5
)

// pastZoom is one fine analysis, and when it was taken.
type pastZoom struct {
	zr ZoomResult
	at float64 // seconds of samples consumed
}

// refZoom is the newest analysis at least refineHop old, and the gap back to it. Nil until the engine
// has been measuring this note that long.
func refZoom(hist []pastZoom, now float64) (*ZoomResult, float64) {
	for i := len(hist) - 1; i >= 0; i-- {
		if gap := now - hist[i].at; gap >= refineHop {
			return &hist[i].zr, gap
		}
	}
	return nil, 0
}

// pushZoom records an analysis and drops the ones too old to refine against.
func pushZoom(hist []pastZoom, zr ZoomResult, now float64) []pastZoom {
	hist = append(hist, pastZoom{zr: zr, at: now})

	keep := 0
	for keep < len(hist) && now-hist[keep].at > maxRefineHop {
		keep++
	}
	return hist[keep:]
}

// How many coarse hops a new note must survive before the tuner switches to it: at ~85 ms a hop, a
// quarter of a second, slower than a transient and faster than a technician moving reeds.
const noteConfirmHops = 3

// How many an OCTAVE JUMP off the tracked note must survive, which is far more because a beating pair
// nulls. Two equal reeds beating at fb cancel outright once every 1/fb: the fundamental is gone and
// its harmonics an octave up (beating at 2*fb, not nulling) are all that is left, so the detector
// reports the octave. The null is short and passes; a real octave does not. 0.85 s outlasts the null
// of any pair the engine will still speak for.
const octaveConfirmHops = 10

// confirmHops: an octave jump waits octaveConfirmHops, everything else noteConfirmHops.
func confirmHops(from, to tuning.Note) int {
	if from != 0 && (to == from+12 || to == from-12) {
		return octaveConfirmHops
	}
	return noteConfirmHops
}

// envelopeBeat reads a beat off the amplitude, but only when the note was sounding throughout the
// window (steady), not starting or dying inside it. How deep the swing must be is left to the caller.
func (e *Engine) envelopeBeat(zr ZoomResult) (hz, depth float64, ok bool) {
	if !e.steady {
		return 0, 0, false
	}
	return EnvelopeBeat(zr, minBeatHz, 25)
}

// splitPair stands the two reeds of a merged pair back up out of its beat.
//
// A pair the spectrum cannot separate comes back as one line, so a single line where more than one
// reed was asked for is the shape this looks for. Three tests, all of which must pass:
//   - the fit will stand behind naming the two reeds (fit.Split): exactly two adjacent reeds by
//     reedPeakFloor, with nothing unnameable a beat from either (see emptyTooth, amCrowded).
//   - the comb accounts for the band (pairExplained), catching a third reed not a whole beat away.
//   - the band's beat is as deep as those two reeds force it (PairDepth, pairDepthSlack).
//
// The fit is returned whether or not it is a pair, because it carries the beat as well as the reeds
// (see beatOf); split is what says the reeds may be reported.
func (e *Engine) splitPair(zr ZoomResult, peaks []Peak, hz float64) (fit PairFit, split bool) {
	if len(peaks) != 1 || hz <= 0 {
		return PairFit{}, false
	}
	fit, ok := FitPair(zr, peaks[0].Freq, hz, e.lobeWidth())
	if !ok {
		return PairFit{}, false
	}
	if !fit.Split || fit.Explained < pairExplained ||
		fit.Depth < PairDepth(fit.Ratio)-pairDepthSlack {
		return fit, false
	}
	return fit, true
}

// beatOf is the beat to put in front of the technician. The envelope's beat is unreliable under a
// working bellows (the stroke and the beat blend in one main lobe), so where the fit stands behind
// the pair its spacing is used; everywhere else the envelope's, as before. Taking the model's spacing
// when it does NOT stand behind the pair was tried and was worse (see the bellows-on-the-beat corner).
func beatOf(fit PairFit, envelopeHz float64) float64 {
	if fit.Split && fit.BeatHz > 0 && fit.Explained >= pairExplained {
		return fit.BeatHz
	}
	return envelopeHz
}

// reedShape is what a fine result claims to have found, and how. Two windows in a row must agree on it
// before it reaches the screen.
type reedShape struct {
	reeds    int
	beats    int
	fromBeat bool
	// Compound mode only: the key the note resolved to and each band's line count, so a base flip
	// or a rank appearing must also repeat before it is reported. Zero in single-band mode. Sized
	// for a six-voice bass machine.
	base  tuning.Note
	found [6]int8
}

// lobeWidth is how much spectrum a single sinusoid occupies under the analysis window: the Hann main
// lobe, 4/T. Two peaks closer than this are one reed's lobe seen twice.
func (e *Engine) lobeWidth() float64 {
	return 4.0 / e.cfg.FineWindow.Seconds()
}
