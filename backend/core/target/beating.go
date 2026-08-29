package target

import (
	"fmt"
	"math"

	"smegg.me/smeggtuner/backend/core/tuning"
)

// SetBeating gives every reed of one note its goal from one beating value in the curve's
// authoring unit: the beat between reeds 1 and 2. Each further pair beats k = 1 + Asymmetry/100
// times the pair below it, the reference's old generator rule, and RefReed sits at pitch.
func (c *Curve) SetBeating(note tuning.Note, value float64) error {
	if !note.Valid() {
		return fmt.Errorf("note %d out of range %d..%d", note, tuning.MinNote, tuning.MaxNote)
	}
	if c.ReedCount < 2 {
		return fmt.Errorf("beating needs at least 2 reeds, the curve has %d", c.ReedCount)
	}
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return fmt.Errorf("beating %g is not a value", value)
	}

	offs := c.beatingOffsets(value)
	for r, off := range offs {
		offs[r] = c.Unit.ToGoal(off, note)
		if !isGoal(offs[r]) {
			return fmt.Errorf("beating %g is not a goal at note %d", value, note)
		}
	}
	copy(c.anchor(note).Reeds, offs)
	return nil
}

// Beating is the beat between reeds 1 and 2 the curve asks for at note, in its authoring unit,
// which round-trips SetBeating; zero for a one-reed or nil curve, or where either has no goal.
func (c *Curve) Beating(note tuning.Note) float64 {
	if c == nil || c.ReedCount < 2 {
		return 0
	}
	goal := c.At(note)
	lo, okLo := goalAt(goal, 0)
	hi, okHi := goalAt(goal, 1)
	if !okLo || !okHi {
		return 0
	}
	return c.Unit.FromGoal(hi, note) - c.Unit.FromGoal(lo, note)
}

// beatingOffsets lays the beats out from the reference reed: beat i is value*k^i, so with three
// reeds beat 2-3 is k times beat 1-2. NoRefReed puts reed 1 at pitch.
func (c *Curve) beatingOffsets(value float64) []float64 {
	n := c.ReedCount
	ref := c.RefReed
	if ref < 0 || ref >= n {
		ref = 0
	}
	k := 1 + c.Asymmetry/MaxAsymmetry

	beats := make([]float64, n-1)
	for i, b := 0, value; i < len(beats); i, b = i+1, b*k {
		beats[i] = b
	}
	out := make([]float64, n)
	for i := ref + 1; i < n; i++ {
		out[i] = out[i-1] + beats[i-1]
	}
	for i := ref - 1; i >= 0; i-- {
		out[i] = out[i+1] - beats[i]
	}
	return out
}
