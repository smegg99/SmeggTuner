package target

import (
	"cmp"
	"slices"

	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// At returns the goal for every slot at note in reference units from that note's scale pitch, and
// meter.Missing where the curve sets none. Between anchors it is linear over the note index
// (Interpolate off takes the nearer anchor, the lower on a tie); past the outermost anchors there
// is no goal unless that end extrapolates; a value above meter.TargetUpperLimit is dropped, as the
// reference target builder drops it. Nil for a nil curve.
func (c *Curve) At(note tuning.Note) []float64 {
	if c == nil {
		return nil
	}
	out := make([]float64, max(c.ReedCount, 0))
	for r := range out {
		out[r] = meter.Missing
	}
	last := len(c.Anchors) - 1
	switch {
	case last < 0:
		return out
	case note < c.Anchors[0].Note:
		if c.ExtrapolateLeft && last > 0 {
			extrapolate(out, note, c.Anchors[0], c.Anchors[1])
		}
	case note > c.Anchors[last].Note:
		if c.ExtrapolateRight && last > 0 {
			extrapolate(out, note, c.Anchors[last], c.Anchors[last-1])
		}
	default:
		c.between(out, note)
	}
	for r, v := range out {
		if v > meter.TargetUpperLimit || v <= meter.Missing {
			out[r] = meter.Missing
		}
	}
	return out
}

// Covers reports whether the curve reaches note: between its outermost anchors, or past an end
// that extrapolates. It is the engine's note coverage; a covered slot may still have no goal.
func (c *Curve) Covers(note tuning.Note) bool {
	if c == nil || len(c.Anchors) == 0 || !note.Valid() {
		return false
	}
	last := len(c.Anchors) - 1
	switch {
	case note < c.Anchors[0].Note:
		return c.ExtrapolateLeft && last > 0
	case note > c.Anchors[last].Note:
		return c.ExtrapolateRight && last > 0
	}
	return true
}

// between fills a note inside the anchors.
func (c *Curve) between(dst []float64, note tuning.Note) {
	i, exact := c.search(note)
	if exact {
		fillFrom(dst, c.Anchors[i])
		return
	}
	lo, hi := c.Anchors[i-1], c.Anchors[i]
	if !c.Interpolate {
		if note-lo.Note <= hi.Note-note {
			fillFrom(dst, lo)
		} else {
			fillFrom(dst, hi)
		}
		return
	}
	t := float64(note-lo.Note) / float64(hi.Note-lo.Note)
	for r := range dst {
		a, b := reedAt(lo, r), reedAt(hi, r)
		dst[r] = a + t*(b-a)
	}
}

// extrapolate continues each reed's line through the end anchor and its neighbour; where the line
// reaches zero it stops, and the goal past that point is zero, as the reference's old generator did.
func extrapolate(dst []float64, note tuning.Note, end, next Anchor) {
	steps := float64(note - end.Note)
	span := float64(next.Note - end.Note)
	for r := range dst {
		a := reedAt(end, r)
		v := a + steps*(reedAt(next, r)-a)/span
		if v*a <= 0 {
			v = 0
		}
		dst[r] = v
	}
}

// search finds note among the anchors: its index if anchored, otherwise the
// index of the first anchor above it, where a new one belongs.
func (c *Curve) search(note tuning.Note) (int, bool) {
	return slices.BinarySearchFunc(c.Anchors, note, func(a Anchor, n tuning.Note) int {
		return cmp.Compare(a.Note, n)
	})
}

func reedAt(a Anchor, reed int) float64 {
	if reed < 0 || reed >= len(a.Reeds) {
		return 0
	}
	return a.Reeds[reed]
}

func fillFrom(dst []float64, a Anchor) {
	for r := range dst {
		dst[r] = reedAt(a, r)
	}
}
