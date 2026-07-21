// Package target holds the goal a tuning is measured against: the per-reed offset
// curve an instrument should hold, and a measurement's error against it.
package target

import (
	"cmp"
	"encoding/json"
	"fmt"
	"math"
	"slices"

	"smegg.me/smeggtuner/core/tuning"
)

// Unit is what a value was authored in. Storage is always cents.
type Unit string

const (
	UnitCents Unit = "cent"
	UnitHz    Unit = "hz"
)

// MinReeds..MaxReeds bound how many reeds a curve may describe.
const (
	MinReeds = 1
	MaxReeds = 8
)

// NoRefReed is RefReed when no reed is defined as being at pitch.
const NoRefReed = -1

// MaxAsymmetry bounds Curve.Asymmetry either way; at the bound the whole beating lies on one side.
const MaxAsymmetry = 100.0

// Anchor is one note the curve is pinned at. Reeds holds cents from that note's
// scale pitch, index 0 for reed 1, and may be shorter than ReedCount.
type Anchor struct {
	Note  tuning.Note `json:"note"`
	Reeds []float64   `json:"reeds"`
}

// Curve is the goal: a sparse set of anchors, interpolated between and held
// flat outside. Zero anchors is legal and gives a pure indicator (zero for
// every reed of every note).
type Curve struct {
	Name      string `json:"name"`
	ReedCount int    `json:"reedCount"` // MinReeds..MaxReeds
	// RefReed is the 0-based reed defined as sounding at pitch, or NoRefReed.
	// Only SetBeating reads it; nothing anywhere may assume which reed it names.
	RefReed int      `json:"refReed"`
	Anchors []Anchor `json:"anchors"` // sorted by Note, at most one per note
	// Unit the anchors were authored in; a display choice - the anchors are always cents.
	Unit Unit `json:"unit"`

	// Asymmetry, in percent (-MaxAsymmetry..MaxAsymmetry), moves the reference
	// reed inside the tremolo without changing the width SetBeating was given.
	Asymmetry float64 `json:"asymmetry"`

	// Interpolate, ExtrapolateLeft and ExtrapolateRight control At between and beyond
	// the anchors. All three default to true; see NewCurve and UnmarshalJSON.
	Interpolate      bool `json:"interpolate"`
	ExtrapolateLeft  bool `json:"extrapolateLeft"`
	ExtrapolateRight bool `json:"extrapolateRight"`
}

func NewCurve(name string, reedCount int) *Curve {
	return &Curve{
		Name:             name,
		ReedCount:        reedCount,
		RefReed:          NoRefReed,
		Unit:             UnitCents,
		Interpolate:      true,
		ExtrapolateLeft:  true,
		ExtrapolateRight: true,
	}
}

// UnmarshalJSON defaults the three interpolation flags to true when absent: on a
// curve saved before those fields existed, missing means "not asked" (yes), not false.
func (c *Curve) UnmarshalJSON(data []byte) error {
	// raw avoids recursion; the pointer fields distinguish a missing flag from a written false.
	type raw Curve
	aux := struct {
		*raw
		Interpolate      *bool `json:"interpolate"`
		ExtrapolateLeft  *bool `json:"extrapolateLeft"`
		ExtrapolateRight *bool `json:"extrapolateRight"`
	}{raw: (*raw)(c)}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	c.Interpolate = aux.Interpolate == nil || *aux.Interpolate
	c.ExtrapolateLeft = aux.ExtrapolateLeft == nil || *aux.ExtrapolateLeft
	c.ExtrapolateRight = aux.ExtrapolateRight == nil || *aux.ExtrapolateRight
	return nil
}

func (c *Curve) Validate() error {
	if c == nil {
		return nil
	}
	if c.ReedCount < MinReeds || c.ReedCount > MaxReeds {
		return fmt.Errorf("reed count %d out of range %d..%d", c.ReedCount, MinReeds, MaxReeds)
	}
	if c.RefReed < NoRefReed || c.RefReed >= c.ReedCount {
		return fmt.Errorf("ref reed %d out of range %d..%d", c.RefReed, NoRefReed, c.ReedCount-1)
	}
	if c.Unit != UnitCents && c.Unit != UnitHz {
		return fmt.Errorf("unknown unit %q", c.Unit)
	}
	if math.IsNaN(c.Asymmetry) || c.Asymmetry < -MaxAsymmetry || c.Asymmetry > MaxAsymmetry {
		return fmt.Errorf("asymmetry %g out of range %g..%g", c.Asymmetry, -MaxAsymmetry, MaxAsymmetry)
	}
	for i, a := range c.Anchors {
		if !a.Note.Valid() {
			return fmt.Errorf("anchor %d: note %d out of range %d..%d",
				i, a.Note, tuning.MinNote, tuning.MaxNote)
		}
		if len(a.Reeds) > c.ReedCount {
			return fmt.Errorf("anchor %d: %d values for %d reeds", i, len(a.Reeds), c.ReedCount)
		}
		if i > 0 && a.Note <= c.Anchors[i-1].Note {
			return fmt.Errorf("anchor %d: note %d not after %d", i, a.Note, c.Anchors[i-1].Note)
		}
	}
	return nil
}

// Sort restores anchor order; Validate still rejects two anchors on one note.
func (c *Curve) Sort() {
	slices.SortStableFunc(c.Anchors, func(a, b Anchor) int { return cmp.Compare(a.Note, b.Note) })
}
