package dsp

import (
	"math"
	"math/cmplx"
)

// The band a hand-worked bellows strokes in. The technician pumps by hand and that stroke is slow;
// outside this band a model given a free frequency finds something to spend it on.
const (
	amMinHz = 0.25
	amMaxHz = 3.0
)

// amGap is how far the bellows must sit from the beat before the two are different things at all. Put
// the stroke on the beat and the upper sideband of the lower reed lands on the upper reed: one line,
// which no fit can split. Inside the gap the bellows is not modelled (the plain comb keeps the
// sideband in the open); a fit that WANTED its bellows there is refused (amCrowded). The value is the
// fit window's resolution, 1/T ~ 0.35 Hz, not a chosen threshold.
const amGap = 0.35

// amClear is how far the fit's bellows must stand off its own beat before the two named reeds can be
// believed - amGap's shadow, wider than the fence. amGap stops the model SAYING something
// unresolvable; it cannot stop the model being PUSHED by it: with the true stroke inside the gap the
// fit widens the comb until the stroke is a legal distance away and reports a spacing that is neither
// beat nor stroke. This must be refused rather than fitted around, because the objective PREFERS the
// lie (a comb 11% narrow can explain more of the band than the truth). Measured by binning splits by
// how far the fit put its bellows from its beat: from 0.45 up matches the model's own noise floor
// (the 288 pairs with no bellows at all); below it the error climbs. It does NOT gate on stroke
// depth: a model that cannot place the stroke under-calls it, so a depth floor would exempt the most
// confused fits. See TestBellowsNearTheBeatRefuses.
const amClear = 0.45

// amCrowded says the fit put its bellows too near its own beat for the two reeds to be told apart.
func amCrowded(fit PairFit) bool {
	if fit.AmHz <= 0 {
		return false
	}
	return math.Abs(fit.AmHz-fit.BeatHz) < amClear
}

// deAM takes the bellows the fit found back out of the band, leaving the reeds swinging only against
// each other. It subtracts the modelled sidebands rather than dividing by a gain (with a coefficient
// per tooth there is no single gain). agrees asks it what the amplitude really swings at; driftRate
// asks how fast the reeds slide, which it could not have asked while the stroke was on them.
func (pb *pairBand) deAM(s amSolution) []complex128 {
	if s.fa <= 0 {
		return pb.band
	}
	out := append([]complex128(nil), pb.band...)
	fk := [combLines]float64{s.f - s.b, s.f, s.f + s.b}
	for k := 0; k < combLines; k++ {
		if cmplx.Abs(s.c[k]) == 0 {
			continue
		}
		up := s.a[k] * s.c[k]
		dn := s.a[k] * cmplx.Conj(s.c[k])
		stepUp := cmplx.Exp(complex(0, 2*math.Pi*(fk[k]+s.fa-pb.center)/pb.rate))
		stepDn := cmplx.Exp(complex(0, 2*math.Pi*(fk[k]-s.fa-pb.center)/pb.rate))
		ru, rd := complex(1, 0), complex(1, 0)
		for i := range out {
			out[i] -= up*ru + dn*rd
			ru *= stepUp
			rd *= stepDn
		}
	}
	return out
}

// amSteps is how finely the stroke rate is swept: the bellows only has to be placed well enough that
// its sidebands land under the ones actually there; the comb decides the reeds, and refine tightens the rate.
const amSteps = 55

// amScanRounds is the alternation budget while the stroke rate is being RANKED (cheaper than settling
// on one: the ordering is stable after two rounds even though the depth is still under-called).
const amScanRounds = 2

// amWorth is how much more of the band the bellows must explain before it is admitted at all. Six real
// numbers always find SOMETHING to explain, and the comb has to shift to let them; a real stroke pays
// easily (its sidebands hold ~m^2/2 of the band's energy), half a percent does not.
const amWorth = 0.005

// findAM looks for the bellows in what the comb could not explain and returns the stroke rate at which
// it best accounts for the leftovers. Zero when there is nothing worth calling a bellows. Each rate is
// scored by one closed-form solve, no re-fitting of the comb, because ranking rates and settling on one are different jobs.
func (pb *pairBand) findAM(f, b float64) float64 {
	plain, ok := pb.solve(f, b, 0, false)
	if !ok {
		return 0
	}
	best, bestScore := 0.0, plain.explained+amWorth
	for i := 0; i <= amSteps; i++ {
		fa := amMinHz + (amMaxHz-amMinHz)*float64(i)/amSteps
		if math.Abs(fa-b) < amGap {
			continue
		}
		s, ok := pb.solveRounds(f, b, fa, false, amScanRounds)
		if !ok {
			continue
		}
		if s.explained > bestScore {
			bestScore, best = s.explained, fa
		}
	}
	return best
}

// combDepth is how deeply the comb's lines swing the amplitude of their own sum, as a fraction of its
// mean: the depth EnvelopeBeat would report if these lines were all there was. The lines beat
// together, so a neighbour reinforces or cancels the pair's beat depending on its PHASE - which is
// what tells a sideband thrown by a pitch modulation from a reed in its own right.
func combDepth(a [combLines]complex128) float64 {
	const steps = 256
	var mean, cos, sin float64
	for i := 0; i < steps; i++ {
		th := 2 * math.Pi * float64(i) / steps
		var sum complex128
		for k, c := range a {
			sum += c * cmplx.Exp(complex(0, float64(k)*th))
		}
		e := cmplx.Abs(sum)
		mean += e
		cos += e * math.Cos(th)
		sin += e * math.Sin(th)
	}
	if mean <= 0 {
		return 0
	}
	// The beat's phase is not ours to choose, so take the magnitude of the first harmonic.
	return 2 * math.Hypot(cos, sin) / mean
}

// edgeHz is how close to the wall of a search window an optimum may land and still count as interior.
const edgeHz = 0.01

// loudestAt is which comb line carries the most, so the bellows is reported as the stroke on the reed
// a technician is actually looking at.
func loudestAt(a [combLines]complex128) int {
	best := 0
	for k := 1; k < combLines; k++ {
		if abs2(a[k]) > abs2(a[best]) {
			best = k
		}
	}
	return best
}
