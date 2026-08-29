package target

import (
	"math"

	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// Unit is what a deviation is written in; the engine, storage and curves always carry reference units.
type Unit string

const (
	UnitReference Unit = "reference" // 1 = 0.06 percent of the note's frequency
	UnitCents     Unit = "cent"
	UnitHz        Unit = "hz"
)

// unitFraction is one reference unit as a fraction of the note's frequency.
const unitFraction = 0.0006

// goalA4 is the A4 a goal typed in Hz is read against, whatever the session's: the reference beat list's rule.
const goalA4 = 440.0

func (u Unit) Valid() bool {
	switch u {
	case UnitReference, UnitCents, UnitHz:
		return true
	}
	return false
}

// FromReference converts d reference units at a note of nominal Hz into u.
func (u Unit) FromReference(d, nominal float64) float64 {
	switch u {
	case UnitCents:
		return 1200 * math.Log2(1+unitFraction*d)
	case UnitHz:
		return meter.FrequencyHz(nominal, d) - nominal
	}
	return d
}

// ToReference converts v, written in u at a note of nominal Hz, into reference units.
func (u Unit) ToReference(v, nominal float64) float64 {
	switch u {
	case UnitCents:
		return (math.Exp2(v/1200) - 1) / unitFraction
	case UnitHz:
		return meter.DeviationUnits(nominal+v, nominal)
	}
	return v
}

// ToGoal converts v, typed in u for note, into the reference units a curve stores. Hz is read at
// the note's pitch for A4 = 440 Hz with the reference's widened 0.06, so a goal never moves with A4.
func (u Unit) ToGoal(v float64, note tuning.Note) float64 {
	if u == UnitHz {
		return v * 100 / (note.Freq(goalA4) * meter.TargetHertzScale)
	}
	return u.ToReference(v, 0)
}

// FromGoal is ToGoal's inverse: d reference units of a goal at note, in u.
func (u Unit) FromGoal(d float64, note tuning.Note) float64 {
	if u == UnitHz {
		return d * note.Freq(goalA4) * meter.TargetHertzScale / 100
	}
	return u.FromReference(d, 0)
}

// isGoal reports whether d reference units can be a goal: a number above the reference's empty marker.
func isGoal(d float64) bool {
	return !math.IsNaN(d) && !math.IsInf(d, 0) && d > meter.Missing
}

// In restates e's deviations in u: cents convert, Hz takes the Hz fields, reference stays.
func (e ReedError) In(u Unit) ReedError {
	switch u {
	case UnitCents:
		e.Curr, e.Goal = UnitCents.FromReference(e.Curr, 0), UnitCents.FromReference(e.Goal, 0)
		e.Error = e.Curr - e.Goal
	case UnitHz:
		e.Curr, e.Goal, e.Error = e.CurrHz, e.GoalHz, e.ErrorHz
	}
	return e
}

// In restates e's beat in u, as ReedError.In does.
func (e BeatError) In(u Unit) BeatError {
	switch u {
	case UnitCents:
		e.Curr, e.Goal = UnitCents.FromReference(e.Curr, 0), UnitCents.FromReference(e.Goal, 0)
		e.Error = e.Curr - e.Goal
	case UnitHz:
		e.Curr, e.Goal, e.Error = e.CurrHz, e.GoalHz, e.ErrorHz
	}
	return e
}
