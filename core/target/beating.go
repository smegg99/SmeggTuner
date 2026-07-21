package target

import (
	"fmt"
	"math"

	"smegg.me/smeggtuner/core/tuning"
)

// SetBeating gives every reed of one note its goal from one beating value in the
// curve's authoring unit (the beat between neighbours), spread by RefReed and Asymmetry.
func (c *Curve) SetBeating(note tuning.Note, value, a4 float64) error {
	if !note.Valid() {
		return fmt.Errorf("note %d out of range %d..%d", note, tuning.MinNote, tuning.MaxNote)
	}
	if c.ReedCount < 2 {
		return fmt.Errorf("beating needs at least 2 reeds, the curve has %d", c.ReedCount)
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("beating %g is not a value", value)
	}

	// One typed step between neighbours; the rank is that step ReedCount-1 times.
	offs := c.beatingOffsets(value * float64(c.ReedCount-1))
	for r, off := range offs {
		if c.Unit == UnitHz {
			offs[r] = CentsFromHz(note, off, a4)
		}
		if math.IsNaN(offs[r]) || math.IsInf(offs[r], 0) {
			return fmt.Errorf("beating %g is not a pitch at note %d", value, note)
		}
	}
	copy(c.anchor(note).Reeds, offs)
	return nil
}

// Beating is the tremolo the curve asks for at note, in its authoring unit: the beat
// between neighbours, round-tripping SetBeating whatever Asymmetry did. Zero for a
// one-reed or nil curve.
func (c *Curve) Beating(note tuning.Note, a4 float64) float64 {
	if c == nil || c.ReedCount < 2 {
		return 0
	}
	goal := c.At(note)
	if len(goal) < 2 {
		return 0
	}
	steps := float64(len(goal) - 1)
	lo, hi := goal[0], goal[len(goal)-1]
	if c.Unit == UnitHz {
		return (HzFromCents(note, hi, a4) - HzFromCents(note, lo, a4)) / steps
	}
	return (hi - lo) / steps
}

// beatingOffsets spreads a beating across the reeds: reference reed at zero, outermost pair exactly value apart; negative mirrors.
func (c *Curve) beatingOffsets(value float64) []float64 {
	n := c.ReedCount
	ref := c.RefReed
	if ref < 0 || ref >= n {
		ref = 0 // NoRefReed: anchor the width on reed 0
	}

	below, above := c.splitBeating(value, ref)
	out := make([]float64, n)
	for i := 0; i < ref; i++ {
		out[i] = -below * float64(ref-i) / float64(ref)
	}
	for i := ref + 1; i < n; i++ {
		out[i] = above * float64(i-ref) / float64(n-1-ref)
	}
	return out
}

// splitBeating divides a beating below/above the reference reed; the two always sum to value.
func (c *Curve) splitBeating(value float64, ref int) (below, above float64) {
	lo := float64(ref) / float64(c.ReedCount-1)
	hi := 1 - lo
	if lo == 0 || hi == 0 {
		// Reference reed at an end: one side has no reeds, nothing to divide.
		return value * lo, value * hi
	}

	t := c.Asymmetry / MaxAsymmetry
	t = math.Max(-1, math.Min(1, t))
	if t >= 0 {
		below = value * lo * (1 - t) // at +1 the whole width is above the reference
		return below, value - below
	}
	above = value * hi * (1 + t) // at -1 the whole width is below it
	return value - above, above
}
