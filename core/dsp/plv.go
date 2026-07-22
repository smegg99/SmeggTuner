package dsp

import (
	"math"
	"math/cmplx"
)

// HarmonicPLV measures how phase-locked band hi is to the k-th harmonic of the line at fLo in band
// lo, as a phase-locking value in 0..1. A partial is the fundamental seen k times faster: it rides
// every wobble the bellows puts on the reed, so against the fundamental's phase times k it holds
// still and the value sits near 1. An independent reed in the hi band drifts on its own phase and
// pulls the value down - unless it beats slower than the window resolves, which is what
// ResidualAngle watches across windows.
//
// Both ZoomResults must come from the same Analyze window (they share their first sample's clock);
// their rates may differ, the lo phase is interpolated. Zero when either band carries too little.
func HarmonicPLV(lo ZoomResult, fLo float64, k int, hi ZoomResult) float64 {
	if !lo.Valid || !hi.Valid || len(lo.TimeSeries) < 64 || len(hi.TimeSeries) < 64 ||
		lo.Rate <= 0 || hi.Rate <= 0 || fLo <= 0 || k < 2 {
		return 0
	}

	// The lo line's phase against its own frequency: demodulate the baseband at fLo-Center and
	// unwrap. What is left is the line's start phase plus its wobble - the part a partial inherits.
	n := len(lo.TimeSeries)
	phase := make([]float64, n)
	prev := 0.0
	for i, z := range lo.TimeSeries {
		t := float64(i) / lo.Rate
		p := cmplx.Phase(z * cmplx.Exp(complex(0, -2*math.Pi*(fLo-lo.Center)*t)))
		for p-prev > math.Pi {
			p -= 2 * math.Pi
		}
		for p-prev < -math.Pi {
			p += 2 * math.Pi
		}
		phase[i] = p
		prev = p
	}

	// Both basebands share t=0, so sample j of hi sits at t=j/hi.Rate in lo's clock too. The
	// predicted partial phase in hi's baseband is k times the lo line's total phase, less the hi
	// heterodyne's own rotation.
	hw := Hann(len(hi.TimeSeries))
	var acc complex128
	var norm float64
	for j, z := range hi.TimeSeries {
		t := float64(j) / hi.Rate
		fi := t * lo.Rate
		i := int(fi)
		if i+1 >= n {
			break
		}
		fr := fi - float64(i)
		phLo := phase[i]*(1-fr) + phase[i+1]*fr
		pred := float64(k)*(phLo+2*math.Pi*fLo*t) - 2*math.Pi*hi.Center*t
		acc += complex(hw[j], 0) * z * cmplx.Exp(complex(0, -pred))
		norm += hw[j] * cmplx.Abs(z)
	}
	if norm <= 0 {
		return 0
	}
	return cmplx.Abs(acc) / norm
}

// ResidualAngle is the hi line's phase against the k-th power of the lo line's, read off the two
// bands' bin phases. The heterodyne restarts at phase zero every Analyze, so a bin's phase tracks
// absolute frequency across hops (the property RefinePhase leans on): a partial's angle holds still
// for as long as the note sounds, an independent line's rotates at exactly the beat - however slow.
// One window cannot tell a partial from a reed beating below its resolution; this angle, watched
// across windows by bandTrack, can - which is how a technician hears a slow beat too, by waiting.
// The k-fold phase ambiguity is harmless: k times a whole turn is a whole turn.
func ResidualAngle(lo ZoomResult, fLo float64, k int, hi ZoomResult, fHi float64) (float64, bool) {
	pl, ok := binPhase(lo, fLo)
	if !ok {
		return 0, false
	}
	ph, ok := binPhase(hi, fHi)
	if !ok {
		return 0, false
	}
	a := math.Mod(ph-float64(k)*pl, 2*math.Pi)
	if a > math.Pi {
		a -= 2 * math.Pi
	} else if a <= -math.Pi {
		a += 2 * math.Pi
	}
	return a, true
}

// binPhase is the line's phase read at the bin nearest f. A symmetric window reports a line's
// phase as of its own centre, so the raw bin phase carries an extra 2*pi*(f-fBin)*tc that
// saw-tooths as drift walks the line across the bin grid - nine radians per hertz at three
// seconds. It is taken off here so the phase follows the line, not the grid.
func binPhase(zr ZoomResult, f float64) (float64, bool) {
	if !zr.Valid || zr.BinHz <= 0 || zr.Rate <= 0 || len(zr.TimeSeries) == 0 {
		return 0, false
	}
	bin := int(math.Round((f - zr.MinHz) / zr.BinHz))
	if bin < 0 || bin >= len(zr.Phases) {
		return 0, false
	}
	fBin := zr.MinHz + float64(bin)*zr.BinHz
	tc := float64(len(zr.TimeSeries)) / (2 * zr.Rate)
	return zr.Phases[bin] - 2*math.Pi*(f-fBin)*tc, true
}
