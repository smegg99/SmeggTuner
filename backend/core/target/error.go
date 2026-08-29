package target

import (
	"math"

	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// DefaultTolerance is the in-tune window when nothing says otherwise: one reference unit.
const DefaultTolerance = 1.0

// DefaultBeatTolerance is wider than the reed window: a beat carries both reeds' errors, so two
// in-tune reeds can sit two units apart. In reference units, since a tremolo is a different Hz at every note.
const DefaultBeatTolerance = 3.0

// Tolerances resolves the two windows a display judges by; a non-positive window falls back to the default.
func Tolerances(reed, beat float64) (float64, float64) {
	if reed <= 0 {
		reed = DefaultTolerance
	}
	if beat <= 0 {
		beat = DefaultBeatTolerance
	}
	return reed, beat
}

// Reading is one measured note, a live measurement or a stored take: its slots and their adjacent
// beats. Note is the instrument's key; Transpose is how many semitones above it the key sounded.
type Reading struct {
	Note      tuning.Note
	Transpose int
	Reeds     []dsp.ReedMeasure
	Beats     []dsp.BeatMeasure
}

// Reference is which of the two error conventions a display follows. They are
// one measurement read two ways: see ReedError.Display.
type Reference string

const (
	// RefScale shows Curr, the deviation from the tempered scale, driven to Goal.
	RefScale Reference = "scale"
	// RefGoal shows Error, the distance from the goal curve, driven to zero.
	RefGoal Reference = "goal"
)

// ReedError is one slot of one note, measured against the goal, in reference units.
type ReedError struct {
	Reed  int     `json:"reed"`  // 0-based slot, as RefReed is: 0 is slot 1
	Curr  float64 `json:"curr"`  // deviation from the slot's scale pitch, measured
	Goal  float64 `json:"goal"`  // deviation the curve asks for; 0 without one
	Error float64 `json:"error"` // Curr - Goal: what has to come off the reed
	// HasGoal says the curve sets a goal for this slot here; without one Goal prints blank.
	HasGoal bool `json:"hasGoal"`

	// The same three in Hz at the slot's own pitch. ErrorHz is the difference of
	// the two deviations, not the deviation of the difference.
	CurrHz  float64 `json:"currHz"`
	GoalHz  float64 `json:"goalHz"`
	ErrorHz float64 `json:"errorHz"`

	InTol bool `json:"inTol"`
}

// Display returns the number to show and the number to drive it to, for the given convention.
func (e ReedError) Display(ref Reference) (show, drive float64) {
	if ref == RefGoal {
		return e.Error, 0
	}
	return e.Curr, e.Goal
}

// Errors joins a reading to the goal curve: one row per present slot, a missing slot has none.
// No goal (no curve, or none at that note and column) gives Goal 0 and Error == Curr; tol defaults
// to DefaultTolerance when not positive.
// cols is the curve column each slot answers to, cols[i] for slot i: a register short of the
// instrument's full one leaves gaps in the card, and without the map a rank is judged against
// another rank's goal. Nil falls back to the slot's position.
func Errors(r Reading, c *Curve, a4, tol float64, cols []int) []ReedError {
	tol, _ = Tolerances(tol, 0)
	goal := c.At(r.Note)

	out := make([]ReedError, 0, len(r.Reeds))
	for i, reed := range r.Reeds {
		if !reed.Present {
			continue
		}
		ref := r.nominal(reed, a4)
		g, has := goalAt(goal, colOf(cols, i))
		currHz, goalHz := reed.Freq-ref, UnitHz.FromReference(g, ref)
		out = append(out, ReedError{
			Reed:    i,
			Curr:    reed.Dev,
			Goal:    g,
			Error:   reed.Dev - g,
			HasGoal: has,
			CurrHz:  currHz,
			GoalHz:  goalHz,
			ErrorHz: currHz - goalHz,
			InTol:   math.Abs(reed.Dev-g) <= tol,
		})
	}
	return out
}

// nominal is the tempered pitch a slot sounds at: the key moved by the transposition and to the slot's own octave.
func (r Reading) nominal(reed dsp.ReedMeasure, a4 float64) float64 {
	return (r.Note + tuning.Note(r.Transpose+reed.Octave)).Freq(a4)
}

// goalAt is column col of an At result, or 0 and false where the curve sets no goal.
func goalAt(goal []float64, col int) (float64, bool) {
	if col < 0 || col >= len(goal) || goal[col] <= meter.Missing {
		return 0, false
	}
	return goal[col], true
}

// colOf is the curve column a slot answers to; an unmapped slot answers to its own position.
func colOf(cols []int, slot int) int {
	if slot >= 0 && slot < len(cols) {
		return cols[slot]
	}
	return slot
}
