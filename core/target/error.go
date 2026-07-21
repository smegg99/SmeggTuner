package target

import (
	"math"

	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/tuning"
)

// DefaultTolerance is the in-tune window when nothing says otherwise: one cent.
const DefaultTolerance = 1.0

// DefaultBeatTolerance is wider than the reed window: a beat carries both reeds'
// errors, so two in-tune reeds can sit two cents apart. In cents not Hz, since a
// tremolo is a different Hz at every note.
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

// Reference is which of the two error conventions a display follows. They are
// one measurement read two ways: see ReedError.Display.
type Reference string

const (
	// RefScale shows Curr, the deviation from the tempered scale, driven to Goal.
	RefScale Reference = "scale"
	// RefGoal shows Error, the distance from the goal curve, driven to zero.
	RefGoal Reference = "goal"
)

// ReedError is one reed of one note, measured against the goal.
type ReedError struct {
	Reed  int     `json:"reed"`  // 0-based, as RefReed is: 0 is reed 1
	Curr  float64 `json:"curr"`  // cents from scale pitch, measured
	Goal  float64 `json:"goal"`  // cents from scale pitch, from the curve; 0 without one
	Error float64 `json:"error"` // Curr - Goal: what has to come off the reed

	// The same three as pitch deviations in Hz at this note, so no display has to
	// convert cents to Hz itself. ErrorHz is the difference of the two deviations,
	// not the deviation of the difference.
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

// Errors joins a measurement to the goal curve: one row per reed it carries. No curve,
// or a curve with no anchors, gives Goal 0 and Error == Curr. Reeds beyond the curve
// get Goal 0. tol defaults to DefaultTolerance when not positive.
//
// Nothing here consults m.ReedsSeparated/m.ReedsFromBeat: when both are false the rows
// are lobes of one merged peak, not reeds, and the caller must check.
func Errors(m dsp.Measurement, c *Curve, a4, tol float64) []ReedError {
	tol, _ = Tolerances(tol, 0)
	goal := c.At(m.Note)
	ref := refPitch(m, a4)

	out := make([]ReedError, 0, len(m.Reeds))
	for i, r := range m.Reeds {
		curr := CurrCents(r, ref)
		g := goalAt(goal, i)
		err := curr - g
		out = append(out, ReedError{
			Reed:    i,
			Curr:    curr,
			Goal:    g,
			Error:   err,
			CurrHz:  hzAt(ref, curr),
			GoalHz:  hzAt(ref, g),
			ErrorHz: hzAt(ref, curr) - hzAt(ref, g),
			InTol:   math.Abs(err) <= tol,
		})
	}
	return out
}

// CurrCents is the deviation a reed sits at, in cents from ref. It recomputes from
// the reed's frequency when there is one (so a pass read back at a different A4 stays
// true), else a hand-edited Curr stands as typed. Every reed reading must come through here.
func CurrCents(r dsp.ReedMeasure, ref float64) float64 {
	if r.Freq > 0 && ref > 0 {
		return tuning.Cents(r.Freq, ref)
	}
	return r.DevCents
}

// refPitch is the pitch m's deviations are measured from: its scale pitch, else the note's pitch at a4.
func refPitch(m dsp.Measurement, a4 float64) float64 {
	if m.ScalePitch > 0 {
		return m.ScalePitch
	}
	return m.Note.Freq(a4)
}

func goalAt(goal []float64, reed int) float64 {
	if reed < 0 || reed >= len(goal) {
		return 0
	}
	return goal[reed]
}

func hzAt(ref, cents float64) float64 { return tuning.FreqAtCents(ref, cents) - ref }
