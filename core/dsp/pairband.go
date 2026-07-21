package dsp

import (
	"math"
	"math/cmplx"
)

// pairBand is the band the model is fitted to, plus the two tables the search runs over. Both tables
// are functions of ONE variable and neither depends on where the comb sits, so both are built once
// and read thousands of times: proj is the band projected onto a single line, gram the inner product
// of two lines a given distance apart. The cubic between two entries is exact to parts in ten million,
// and the answer the caller finally reads is solved on the samples anyway.
type pairBand struct {
	band   []complex128
	w2     []float64
	rate   float64
	center float64
	energy float64

	projLo, projStep float64
	projTab          []complex128
	gramLo, gramStep float64
	gramTab          []complex128
}

// tabStep is the spacing of both tables, in Hz.
const tabStep = 0.02

func newPairBand(zr ZoomResult, guessHz, searchHz, bLo, bHi, faMax, chirp float64) *pairBand {
	// The cut has to clear the outermost line the model can ask about and stay far below the lowest
	// fundamental a pair can merge at, so the reed's own second harmonic is still taken off.
	cut := math.Max(4*bHi, bHi+faMax+2)
	// Copied, because the de-chirp below rewrites it in place and bandLimit may hand back the caller's slice.
	band := append([]complex128(nil), bandLimit(zr.TimeSeries, zr.Rate, math.Max(cut, 6))...)
	n := len(band)
	if n < 64 {
		return nil
	}

	// The Hann window is applied to the model as well as the data, so it cannot bias the frequencies:
	// it only says which samples the fit listens to hardest. Zoom hands back TimeSeries unwindowed.
	hw := Hann(n)
	pb := &pairBand{
		band:   band,
		w2:     make([]float64, n),
		rate:   zr.Rate,
		center: zr.Center,
	}
	var sw float64
	for i := 0; i < n; i++ {
		pb.w2[i] = hw[i] * hw[i]
		sw += pb.w2[i]
	}
	if sw <= 0 {
		return nil
	}

	// Take the note's own slide out, so the model can go on being lines. A sliding line is a chirp,
	// smeared across as much spectrum as it slid, and a mis-specified model leaks what it cannot
	// account for into the bellows. Both reeds slide together, so the beat never changes and the
	// frequencies come back unchanged in the mean (time is measured from the middle of the window).
	// driftRate measures the rate.
	if chirp != 0 {
		for i := range pb.band {
			t := pb.at0(i)
			pb.band[i] *= cmplx.Exp(complex(0, -math.Pi*chirp*t*t))
		}
	}

	for i := 0; i < n; i++ {
		z := pb.band[i]
		pb.energy += pb.w2[i] * (real(z)*real(z) + imag(z)*imag(z))
	}
	if pb.energy <= 0 {
		return nil
	}
	// Every line the search can place sits within this of the guess: a centre up to searchHz away, a
	// tooth up to bHi beyond, and a bellows sideband faMax beyond that.
	reach := searchHz + bHi + faMax + 2*tabStep

	pb.projStep = tabStep
	pb.projLo = guessHz - reach
	np := int(2*reach/tabStep) + 4
	pb.projTab = make([]complex128, np)
	for i := range pb.projTab {
		pb.projTab[i] = pb.projAt(pb.projLo + float64(i)*tabStep)
	}

	// And no two of those lines are further apart than this. The table starts below zero because the
	// cubic wants a sample either side, and two lines on top of each other is the commonest lookup.
	span := 2*(bHi+faMax) + 2*tabStep
	pb.gramStep = tabStep
	pb.gramLo = -2 * tabStep
	ng := int((span-pb.gramLo)/tabStep) + 4
	pb.gramTab = make([]complex128, ng)
	for i := range pb.gramTab {
		pb.gramTab[i] = pb.gramAt(pb.gramLo + float64(i)*tabStep)
	}
	return pb
}

// driftRate measures how fast the note is sliding, in hertz per second, for newPairBand to de-chirp by.
//
// Not by regressing instantaneous frequency against time: two reeds beating drag that back and forth
// once a beat, and the leftover leans on the slope. The beat can be cancelled exactly instead. Weight
// the instantaneous frequency by the energy carrying it and two tones give |z|^2 * f(t) = M(t) *
// |z|^2(t) + D, where D is a CONSTANT, so the slide fits against |z|^2, t*|z|^2 and one. On one
// condition: no bellows (under a stroke D becomes D*u^2). So FitPair fits once to find the stroke,
// divides it out, and measures the slide against what is left.
func driftRate(band []complex128, w2 []float64, rate float64) (float64, bool) {
	var m [3][3]float64
	var v [3]float64
	for i := 0; i+1 < len(band); i++ {
		z0, z1 := band[i], band[i+1]
		// Im(z1 * conj(z0)) * rate / 2pi is |z|^2 times the frequency in hertz.
		d := z1 * cmplx.Conj(z0)
		nu := imag(d) * rate / (2 * math.Pi)
		e := real(z0)*real(z0) + imag(z0)*imag(z0)
		t := (float64(i) - float64(len(band))/2) / rate
		b := [3]float64{e, t * e, 1}
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				m[j][k] += w2[i] * b[j] * b[k]
			}
			v[j] += w2[i] * b[j] * nu
		}
	}
	x, ok := solve3(m, v)
	if !ok || math.IsNaN(x[1]) || math.Abs(x[1]) > maxDriftHzPerSec {
		return 0, false
	}
	return x[1], true
}

// maxDriftHzPerSec bounds what will be believed as a reed sliding under the bellows. The sweep slides
// a pair 0.1 Hz/s; past half a hertz a second is a note arriving or leaving.
const maxDriftHzPerSec = 0.5

// at0 is sample i's time, measured from the middle of the window.
func (pb *pairBand) at0(i int) float64 {
	return (float64(i) - float64(len(pb.band))/2) / pb.rate
}
