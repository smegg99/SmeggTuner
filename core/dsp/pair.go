package dsp

import (
	"math"
	"math/cmplx"
)

// combLines is 3, not 2: a two-line fit cannot separate a merged pair from one reed whose bellows
// swings it, because a stroke reinforces one sideband and cancels the other, leaving a carrier and
// one companion. What tells them apart is the OTHER side of the carrier, a beat down, and a two-line
// model has nowhere to put it. Three lines represent all four things the band can be exactly - a
// pair, a reed with both sidebands, a reed with one, and three merged reeds - so the fit never guesses.
const combLines = 3

// PairFit is what a comb of lines a beat apart found in the band.
type PairFit struct {
	// Lo and Hi are the two reed frequencies, ascending, exactly BeatHz apart. Only meaningful when Split.
	Lo, Hi float64
	// BeatHz is the spacing the comb settled on, which is NOT in general the beat it was given (see
	// FitPair). Report this one. Meaningful whenever Reeds is 2.
	BeatHz float64
	// Reeds is how many comb lines cleared reedPeakFloor. Two adjacent lines is a pair (the only shape
	// this package stands a reed split on); one is a reed with bellows either side; three is a musette
	// that merged whole.
	Reeds int
	// Split says the fit will stand behind NAMING the two reeds - stricter than Reeds being 2, because
	// the beat outlives the reeds: a bellows stroking near the beat rate lands a sideband on a reed, so
	// the reeds cannot be named (see emptyTooth, amCrowded) but the beat is untouched.
	Split bool
	// Ratio is the quieter reed's amplitude over the louder one's, 0..1.
	Ratio float64
	// Depth is how deeply the comb's lines swing the amplitude of their sum at the beat, weighed
	// against PairDepth(Ratio). Rebuilt from the fitted lines rather than the raw amplitude, because
	// below ~E2 the reed's second harmonic survives decimation and fills the beat's nulls.
	Depth float64
	// Explained is the fraction of the band's energy the model accounts for - the comb plus, when
	// there was one, the bellows working it. Catches a third reed the comb has no tooth for.
	Explained float64
	// LoLouder says which side the louder reed is on, which the beat alone cannot.
	LoLouder bool
	// AmHz and AmDepth are the common-mode amplitude modulation found on the band: the bellows, at its
	// stroke rate, as a fraction of the carrier it swings. AmHz is 0 when nothing was modelled.
	AmHz, AmDepth float64
	// ThirdTooth is the loudest comb line that is NOT one of the two reeds, over the loudest that is:
	// the ground between too loud to be nothing and too quiet to be named. Only meaningful when Reeds is 2.
	ThirdTooth float64
}

// FitPair recovers the two reed frequencies of a merged pair.
//
// Below middle C a pair sits inside the window's main lobe, and the composite peak the picker returns
// is not the pair's centre (measured within 0.005 Hz of the LOWER reed on a 1.2 Hz pair at C3). The
// pair is recoverable anyway, because two Hann-windowed lines 0.6-2 Hz apart over three seconds are
// nearly orthogonal, so a model of lines at known spacing fits them apart:
//
//	z(t) ~= u(t) * [ a0*exp(i2pi(f-b)t) + a1*exp(i2pi*f*t) + a2*exp(i2pi(f+b)t) ]
//
// The spacing is searched over a lobe either side of the beat the envelope read (the envelope blends
// the beat with the bellows stroke and cannot separate them; the band can, because there the two are
// different SHAPES). u(t) = 1 + 2*Re(c*exp(i2pi*fa*t)) is the bellows: one real gain common to the
// band. Two guards: the bellows may not sit within amGap of the beat, nor claim a sideband louder
// than reedPeakFloor of what it hangs off. Nothing here decides whether the pair is REAL - Reeds,
// Depth and Explained are what the caller judges it by.
func FitPair(zr ZoomResult, guessHz, beatHz, searchHz float64) (PairFit, bool) {
	if !zr.Valid || len(zr.TimeSeries) < 256 || zr.Rate <= 0 || beatHz <= 0 || searchHz <= 0 {
		return PairFit{}, false
	}

	// A beat this fast is not a merged pair: the reeds are lobes apart and the picker found both. No
	// blend to undo, so the spacing stands and the fit is the plain comb.
	amAware := beatHz <= 3*searchHz

	bLo, bHi := beatHz, beatHz
	faMax := 0.0
	if amAware {
		bLo = math.Max(minBeatHz, beatHz-searchHz)
		bHi = beatHz + searchHz
		faMax = amMaxHz
	}

	// Twice, and the second pass is not a refinement: it is the same fit against a band the first
	// cleaned up. The reeds slide while they sound (see newPairBand), and the slide cannot be measured
	// while the bellows is still on the band. So: fit once to find the bellows, divide it out, measure
	// the slide against what is left, fit again on a band with neither.
	pb := newPairBand(zr, guessHz, searchHz, bLo, bHi, faMax, 0)
	if pb == nil {
		return PairFit{}, false
	}
	best, ok := pb.search(guessHz, searchHz, bLo, bHi, amAware)
	if !ok {
		return PairFit{}, false
	}
	if amAware {
		// A slide too slow to smear a line (under a fiftieth of a hertz across the window) is not worth refitting for.
		if rate, ok := driftRate(pb.deAM(best), pb.w2, pb.rate); ok && math.Abs(rate) > 0.007 {
			if pb2 := newPairBand(zr, guessHz, searchHz, bLo, bHi, faMax, rate); pb2 != nil {
				if best2, ok := pb2.search(guessHz, searchHz, bLo, bHi, amAware); ok {
					pb, best = pb2, best2
				}
			}
		}
	}

	// An optimum against the wall of the search window is not what the spectrum saw.
	if math.Abs(best.f-guessHz) >= searchHz*0.999 {
		return PairFit{}, false
	}
	if amAware && (best.b <= bLo+edgeHz || best.b >= bHi-edgeHz) {
		return PairFit{}, false
	}

	// The last solve runs on the band itself rather than the search tables, so nothing the caller reads is interpolated.
	final, ok := pb.solve(best.f, best.b, best.fa, true)
	if !ok {
		return PairFit{}, false
	}
	a := final.a

	// Which comb lines are reeds: the same floor the peak picker applies to any line.
	var loudest float64
	for _, c := range a {
		if v := cmplx.Abs(c); v > loudest {
			loudest = v
		}
	}
	if loudest <= 0 {
		return PairFit{}, false
	}
	fit := PairFit{
		BeatHz:    final.b,
		Depth:     combDepth(a),
		Explained: final.explained,
		AmDepth:   2 * cmplx.Abs(final.c[loudestAt(final.a)]),
	}
	if fit.AmDepth > 0 {
		fit.AmHz = final.fa
	}
	var reeds []int
	for k, c := range a {
		if cmplx.Abs(c)/loudest >= reedPeakFloor {
			reeds = append(reeds, k)
		}
	}
	fit.Reeds = len(reeds)
	if len(reeds) != 2 || reeds[1] != reeds[0]+1 {
		return fit, true // not a pair, and the caller is told how many it is
	}

	lo, hi := a[reeds[0]], a[reeds[1]]
	loAmp, hiAmp := cmplx.Abs(lo), cmplx.Abs(hi)
	louder, quieter := loAmp, hiAmp
	if hiAmp > loAmp {
		louder, quieter = hiAmp, loAmp
	}
	fit.Lo = final.f + float64(reeds[0]-combLines/2)*final.b
	fit.Hi = final.f + float64(reeds[1]-combLines/2)*final.b
	fit.Ratio = quieter / louder
	fit.LoLouder = loAmp >= hiAmp
	for k, c := range a {
		if k != reeds[0] && k != reeds[1] {
			fit.ThirdTooth = cmplx.Abs(c) / louder
		}
	}
	if amCrowded(fit) || fit.ThirdTooth > emptyTooth {
		// Something a beat from a reed the model cannot name: a sideband in the tooth a pair leaves
		// empty, or a bellows too close to the beat. This fit cannot name the reeds, and neither can
		// anything in three seconds. The BEAT still stands.
		fit.Lo, fit.Hi, fit.Ratio = 0, 0, 0
		return fit, true
	}
	fit.Split = true
	return fit, true
}

// emptyTooth is how much a pair may leave in the comb tooth that is not one of its reeds. A pair
// leaves it empty (measured: nine in ten under 0.014). What fills it is a bellows stroking at the beat
// - its lower sideband lands here at 13-28% of the carrier - so 0.06 sits in a wide gap between the
// two populations. Above it the split is refused; the beat still goes out.
const emptyTooth = 0.06
