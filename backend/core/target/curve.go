// Package target holds the goal a tuning is measured against: the per-reed offset
// curve an instrument should hold, and a measurement's error against it.
package target

import (
	"cmp"
	"fmt"
	"math"
	"slices"

	"smegg.me/smeggtuner/backend/core/tuning"
)

// MinReeds..MaxReeds bound how many reeds a curve may describe.
const (
	MinReeds = 1
	MaxReeds = 8
)

// NoRefReed is RefReed when no reed is defined as being at pitch.
const NoRefReed = -1

// MaxAsymmetry bounds Curve.Asymmetry either way: -100 stills beat 2-3, +100 doubles it.
const MaxAsymmetry = 100.0

// Anchor is one note the curve is pinned at. Reeds holds reference units from that note's
// scale pitch, index 0 for slot 1; it may be shorter than ReedCount, and a reed past its end reads 0.
type Anchor struct {
	Note  tuning.Note `json:"note"`
	Reeds []float64   `json:"reeds"`
}

// Curve is the goal: a sparse set of anchors, interpolated between, with no goal outside them
// unless an end extrapolates. Zero anchors is legal and gives a pure indicator (no goal anywhere).
type Curve struct {
	Name      string `json:"name"`
	ReedCount int    `json:"reedCount"` // MinReeds..MaxReeds
	// RefReed is the 0-based reed defined as sounding at pitch, or NoRefReed.
	// Only SetBeating reads it; nothing anywhere may assume which reed it names.
	RefReed int      `json:"refReed"`
	Anchors []Anchor `json:"anchors"` // sorted by Note, at most one per note
	// Unit the anchors are typed in (see ToGoal); a display choice, the anchors are always reference units.
	Unit Unit `json:"unit"`

	// Asymmetry, in percent (-MaxAsymmetry..MaxAsymmetry), makes each pair SetBeating writes beat
	// 1 + Asymmetry/100 times the pair below it.
	Asymmetry float64 `json:"asymmetry"`

	// Interpolate, ExtrapolateLeft and ExtrapolateRight control At between and beyond the anchors.
	// NewCurve interpolates and does not extrapolate, as the reference beat list does.
	Interpolate      bool `json:"interpolate"`
	ExtrapolateLeft  bool `json:"extrapolateLeft"`
	ExtrapolateRight bool `json:"extrapolateRight"`
}

func NewCurve(name string, reedCount int) *Curve {
	return &Curve{
		Name:        name,
		ReedCount:   reedCount,
		RefReed:     NoRefReed,
		Unit:        UnitReference,
		Interpolate: true,
	}
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
	if !c.Unit.Valid() {
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
