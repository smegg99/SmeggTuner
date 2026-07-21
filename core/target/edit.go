package target

import (
	"fmt"
	"math"
	"slices"

	"smegg.me/smeggtuner/core/tuning"
)

// Set gives one reed of one note a goal, keeping the anchors sorted. value is in
// the curve's authoring unit; storage stays canonical in cents.
func (c *Curve) Set(note tuning.Note, reed int, value, a4 float64) error {
	if !note.Valid() {
		return fmt.Errorf("note %d out of range %d..%d", note, tuning.MinNote, tuning.MaxNote)
	}
	if reed < 0 || reed >= c.ReedCount {
		return fmt.Errorf("reed %d out of range 0..%d", reed, c.ReedCount-1)
	}
	cents := value
	if c.Unit == UnitHz {
		cents = CentsFromHz(note, value, a4)
	}
	if math.IsNaN(cents) || math.IsInf(cents, 0) {
		return fmt.Errorf("value %g is not a pitch at note %d", value, note)
	}
	c.anchor(note).Reeds[reed] = cents
	return nil
}

// Clear drops the anchor at note. Clearing an unanchored note is a no-op.
func (c *Curve) Clear(note tuning.Note) {
	if i, exact := c.search(note); exact {
		c.Anchors = slices.Delete(c.Anchors, i, i+1)
	}
}

// anchor returns the anchor for note, inserting it in sort order if it is new.
func (c *Curve) anchor(note tuning.Note) *Anchor {
	i, exact := c.search(note)
	if exact {
		a := &c.Anchors[i]
		for len(a.Reeds) < c.ReedCount {
			a.Reeds = append(a.Reeds, 0)
		}
		return a
	}
	c.Anchors = append(c.Anchors, Anchor{})
	copy(c.Anchors[i+1:], c.Anchors[i:])
	c.Anchors[i] = Anchor{Note: note, Reeds: make([]float64, c.ReedCount)}
	return &c.Anchors[i]
}
