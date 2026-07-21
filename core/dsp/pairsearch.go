package dsp

import (
	"math"
)

// search sweeps the spacing, and the centre and stroke rate with it, and returns the placement that
// best accounts for the band. For each candidate spacing the comb is placed where it best explains the
// band, the bellows is looked for in what the comb could not, and the two are solved against each
// other. The score is the fraction of the band explained. A plain comb cannot make this choice - it is
// flat and wrong across spacings - which is why the bellows is in the model at all.
func (pb *pairBand) search(guessHz, searchHz, bLo, bHi float64, amAware bool) (amSolution, bool) {
	steps := 0
	if amAware {
		steps = beatSweep
	}
	var best amSolution
	bestScore := math.Inf(-1)
	for i := 0; i <= steps; i++ {
		b := bLo
		if steps > 0 {
			b = bLo + (bHi-bLo)*float64(i)/float64(steps)
		}
		s, ok := pb.at(b, guessHz, searchHz, amAware)
		if !ok || s.explained <= bestScore {
			continue
		}
		// The spacing has to be the one the amplitude actually swings at (see agrees).
		if amAware && !pb.agrees(s) {
			continue
		}
		bestScore, best = s.explained, s
	}
	if math.IsInf(bestScore, -1) {
		return best, false // no spacing the band will stand behind
	}

	// Walk up the hill the sweep found, one parameter at a time. The score is not a parabola (comb
	// sidelobes, two equally good centres for a pair), so the sweep finds the hill and this climbs it.
	best = pb.refine(best, guessHz, searchHz, bLo, bHi, amAware)
	if amAware && !pb.agrees(best) {
		return best, false // the refinement walked it off the beat
	}
	return best, true
}

// at places the comb at spacing b: the centre where the band is best accounted for, then the bellows in what is left.
func (pb *pairBand) at(b, guessHz, searchHz float64, amAware bool) (amSolution, bool) {
	if b <= 0 {
		return amSolution{}, false
	}
	f, ok := pb.combCentre(b, guessHz, searchHz)
	if !ok {
		return amSolution{}, false
	}
	fa := 0.0
	if amAware {
		fa = pb.findAM(f, b)
	}
	return pb.solve(f, b, fa, false)
}

// beatAgree is how far the comb's spacing may sit from the beat the amplitude actually swings at, once
// the bellows has been divided out. A tenth of a hertz is several times the envelope's own bin
// spacing and well inside the spread of an imperfect de-AM.
const beatAgree = 0.10

// agrees asks whether the spacing the comb settled on is the spacing the band's amplitude swings at.
// Without it the search finds two kinds of nonsense that explain the ENERGY better than the truth: the
// half-beat alias (a comb of spacing b/2 has a spare line to mop up noise) and, where the bellows
// strokes at half the beat, the two sidebands that add midway between the reeds. Energy cannot
// separate these; the amplitude can, so the model divides its OWN bellows out and answers for the rest.
func (pb *pairBand) agrees(s amSolution) bool {
	clean := pb.deAM(s)
	if clean == nil {
		return false
	}
	hz, _, ok := EnvelopeBeat(ZoomResult{
		Valid:      true,
		Rate:       pb.rate,
		Center:     pb.center,
		TimeSeries: clean,
	}, minBeatHz, 25)
	if !ok {
		// Nothing swinging at all: not a beat, but the caller's guards read the lines, so let them.
		return true
	}
	return math.Abs(hz-s.b) <= beatAgree
}

// combCentre sweeps the window for the centre at which a plain comb of spacing b best accounts for the
// band, then walks up the hill. Plain, because the bellows cannot be looked for until there is a comb
// to hang it off, and refine puts back the hundredths of a hertz the sidebands pull.
func (pb *pairBand) combCentre(b, guessHz, searchHz float64) (float64, bool) {
	score := func(f float64) float64 {
		s, ok := pb.solve(f, b, 0, false)
		if !ok {
			return math.Inf(-1)
		}
		return s.explained
	}
	step := 2 * searchHz / combSweep
	bestAt, best, bestScore := -1, 0.0, math.Inf(-1)
	for i := 0; i <= combSweep; i++ {
		f := guessHz - searchHz + float64(i)*step
		if s := score(f); s > bestScore {
			bestAt, best, bestScore = i, f, s
		}
	}
	if bestAt < 0 {
		return 0, false
	}
	return goldenMax(score, best-step, best+step), true
}

// How many places in the window the centre and the spacing are each tried at.
const (
	combSweep = 48
	beatSweep = 48
)

// refine walks up the hill the sweep found, one parameter at a time: centre, spacing, stroke rate,
// twice round. Coordinate-wise and not jointly, because the three are nearly separable once the model
// is roughly right, and a golden section on each is a dozen evaluations against hundreds for a joint search.
func (pb *pairBand) refine(s amSolution, guessHz, searchHz, bLo, bHi float64, amAware bool) amSolution {
	score := func(f, b, fa float64) float64 {
		if b <= 0 {
			return math.Inf(-1)
		}
		n, ok := pb.solve(f, b, fa, false)
		if !ok {
			return math.Inf(-1)
		}
		return n.explained
	}
	// One grid step either side of where the sweep left each parameter.
	fStep := 2 * searchHz / combSweep
	bStep := (bHi - bLo) / beatSweep
	faStep := (amMaxHz - amMinHz) / amSteps

	take := func(f, b, fa float64) {
		if score(f, b, fa) > score(s.f, s.b, s.fa) {
			s.f, s.b, s.fa = f, b, fa
		}
	}
	for round := 0; round < 2; round++ {
		take(goldenMax(func(x float64) float64 { return score(x, s.b, s.fa) },
			math.Max(guessHz-searchHz, s.f-fStep), math.Min(guessHz+searchHz, s.f+fStep)),
			s.b, s.fa)
		if !amAware {
			break
		}
		take(s.f, goldenMax(func(x float64) float64 { return score(s.f, x, s.fa) },
			math.Max(bLo, s.b-bStep), math.Min(bHi, s.b+bStep)), s.fa)
		if s.fa <= 0 {
			continue
		}
		lo := math.Max(amMinHz, s.fa-faStep)
		hi := math.Min(amMaxHz, s.fa+faStep)
		// The refinement may not walk the bellows onto the beat either. See amGap.
		if s.fa > s.b {
			lo = math.Max(lo, s.b+amGap)
		} else {
			hi = math.Min(hi, s.b-amGap)
		}
		if hi > lo {
			take(s.f, s.b, goldenMax(func(x float64) float64 { return score(s.f, s.b, x) }, lo, hi))
		}
	}
	if n, ok := pb.solve(s.f, s.b, s.fa, false); ok {
		return n
	}
	return s
}
