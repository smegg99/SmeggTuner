package target

import (
	"math"
	"strconv"
	"strings"

	"smegg.me/smeggtuner/backend/core/dsp/meter"
)

// BeatError is one adjacent slot pair of one note, measured against the curve's beat goal.
type BeatError struct {
	Pair string `json:"pair"` // "1-2", "2-3", as dsp.BeatMeasure spells it
	Low  int    `json:"low"`  // 0-based slots, lower first
	High int    `json:"high"`

	Curr  float64 `json:"curr"`  // the measured beat in reference units at the upper slot's pitch
	Goal  float64 `json:"goal"`  // the upper slot's goal minus the lower's; 0 without both goals
	Error float64 `json:"error"` // Curr - Goal
	// HasGoal says both slots have a goal here, so the pair has one.
	HasGoal bool `json:"hasGoal"`

	CurrHz  float64 `json:"currHz"`
	GoalHz  float64 `json:"goalHz"`
	ErrorHz float64 `json:"errorHz"`

	InTol bool `json:"inTol"` // |Error| <= tolerance, in reference units
}

// BeatErrors joins a reading's measured beats to the goal curve; a pair touching a missing slot
// has none. tol defaults to DefaultBeatTolerance when not positive; cols as in Errors.
func BeatErrors(r Reading, c *Curve, a4, tol float64, cols []int) []BeatError {
	_, tol = Tolerances(0, tol)
	goal := c.At(r.Note)

	out := make([]BeatError, 0, len(r.Beats))
	for _, b := range r.Beats {
		lo, hi, ok := parsePair(b.Pair)
		if !ok || !b.Present || hi >= len(r.Reeds) || !r.Reeds[lo].Present || !r.Reeds[hi].Present {
			continue
		}
		ref := r.nominal(r.Reeds[hi], a4) // an octave pair's beat is the deviations' difference at the upper pitch
		low, okLow := goalAt(goal, colOf(cols, lo))
		high, okHigh := goalAt(goal, colOf(cols, hi))
		has := okLow && okHigh
		curr := meter.DeviationUnits(ref+b.Hz, ref)
		g, goalHz := 0.0, 0.0
		if has {
			g, goalHz = meter.TargetPair(low, high, ref, true, false), meter.TargetPair(low, high, ref, true, true)
		}
		out = append(out, BeatError{
			Pair:    b.Pair,
			Low:     lo,
			High:    hi,
			Curr:    curr,
			Goal:    g,
			Error:   curr - g,
			HasGoal: has,
			CurrHz:  b.Hz,
			GoalHz:  goalHz,
			ErrorHz: b.Hz - goalHz,
			InTol:   math.Abs(curr-g) <= tol,
		})
	}
	return out
}

// parsePair reads dsp.BeatMeasure.Pair, which is 1-based, into slot indices.
func parsePair(pair string) (lo, hi int, ok bool) {
	a, b, found := strings.Cut(pair, "-")
	if !found {
		return 0, 0, false
	}
	l, err := strconv.Atoi(a)
	if err != nil {
		return 0, 0, false
	}
	h, err := strconv.Atoi(b)
	if err != nil || l < 1 || h <= l {
		return 0, 0, false
	}
	return l - 1, h - 1, true
}
