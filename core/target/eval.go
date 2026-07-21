package target

import (
	"cmp"
	"slices"

	"smegg.me/smeggtuner/core/tuning"
)

// At returns the goal for every reed at note, in cents from that note's scale
// pitch. Between anchors it interpolates linearly; past the outermost anchor it
// holds flat. Interpolate off takes the nearer anchor's value (the lower on a
// tie); an Extrapolate flag off gives zero past that end instead of holding
// flat. An anchor is never extrapolation. A nil or anchorless curve gives zeros.
func (c *Curve) At(note tuning.Note) []float64 {
	if c == nil {
		return nil
	}
	n := c.ReedCount
	if n < 0 {
		n = 0
	}
	out := make([]float64, n)
	if len(c.Anchors) == 0 {
		return out
	}

	first := c.Anchors[0]
	last := c.Anchors[len(c.Anchors)-1]
	if note <= first.Note {
		if note < first.Note && !c.ExtrapolateLeft {
			return out
		}
		fillFrom(out, first)
		return out
	}
	if note >= last.Note {
		if note > last.Note && !c.ExtrapolateRight {
			return out
		}
		fillFrom(out, last)
		return out
	}

	i, exact := c.search(note)
	if exact {
		fillFrom(out, c.Anchors[i])
		return out
	}
	hi := c.Anchors[i]
	lo := c.Anchors[i-1]
	if !c.Interpolate {
		if note-lo.Note <= hi.Note-note {
			fillFrom(out, lo)
		} else {
			fillFrom(out, hi)
		}
		return out
	}
	t := float64(note-lo.Note) / float64(hi.Note-lo.Note)
	for r := range out {
		a, b := reedAt(lo, r), reedAt(hi, r)
		out[r] = a + t*(b-a)
	}
	return out
}

// search finds note among the anchors: its index if anchored, otherwise the
// index of the first anchor above it, where a new one belongs.
func (c *Curve) search(note tuning.Note) (int, bool) {
	return slices.BinarySearchFunc(c.Anchors, note, func(a Anchor, n tuning.Note) int {
		return cmp.Compare(a.Note, n)
	})
}

// CentsFromHz converts an Hz deviation from note's scale pitch at a4 into cents.
func CentsFromHz(note tuning.Note, hz, a4 float64) float64 {
	f0 := note.Freq(a4)
	return tuning.Cents(f0+hz, f0)
}

// HzFromCents is the inverse: the Hz a cents offset amounts to at this note.
func HzFromCents(note tuning.Note, cents, a4 float64) float64 {
	return hzAt(note.Freq(a4), cents)
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
