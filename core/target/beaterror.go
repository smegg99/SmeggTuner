package target

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/tuning"
)

// BeatError is one pair of reeds of one note, measured against the curve's beat goal.
type BeatError struct {
	Pair string `json:"pair"` // "1-2", "1-3", "2-3", as dsp.BeatMeasure spells it
	Low  int    `json:"low"`  // 0-based reed indices, lower pitch first
	High int    `json:"high"`

	Curr  float64 `json:"curr"`  // cents between the two reeds, measured
	Goal  float64 `json:"goal"`  // cents between their two goals; 0 without a curve
	Error float64 `json:"error"` // Curr - Goal

	CurrHz  float64 `json:"currHz"`
	GoalHz  float64 `json:"goalHz"`
	ErrorHz float64 `json:"errorHz"`

	InTol bool `json:"inTol"` // |Error| <= tolerance, in cents
	// FromEnvelope marks a beat read off the amplitude when the two reeds do not separate.
	FromEnvelope bool `json:"fromEnvelope"`
}

// BeatErrors joins a measurement's beats to the goal curve. Reeds separated, it
// returns every pair in reed order (1-2, 1-3, 2-3, ...); not separated, only the
// envelope beats, since per-reed frequencies are unusable there - the inverse of
// the ReedError rule. tol defaults to DefaultBeatTolerance when not positive.
func BeatErrors(m dsp.Measurement, c *Curve, a4, tol float64) []BeatError {
	_, tol = Tolerances(0, tol)
	ref := refPitch(m, a4)
	if ref <= 0 {
		return nil
	}
	goal := c.At(m.Note)

	row := func(lo, hi int, curr float64, fromEnvelope bool) BeatError {
		g := goalAt(goal, hi) - goalAt(goal, lo)
		err := curr - g
		return BeatError{
			Pair:         fmt.Sprintf("%d-%d", lo+1, hi+1),
			Low:          lo,
			High:         hi,
			Curr:         curr,
			Goal:         g,
			Error:        err,
			CurrHz:       hzAt(ref, curr),
			GoalHz:       hzAt(ref, g),
			ErrorHz:      hzAt(ref, curr) - hzAt(ref, g),
			InTol:        math.Abs(err) <= tol,
			FromEnvelope: fromEnvelope,
		}
	}

	if !m.ReedsSeparated {
		out := make([]BeatError, 0, len(m.Beats))
		for _, b := range m.Beats {
			lo, hi, ok := parsePair(b.Pair)
			if !ok {
				continue
			}
			// Envelope beat is a positive rate; the goal runs low reed to high, positive on an ascending curve.
			out = append(out, row(lo, hi, tuning.Cents(ref+b.Hz, ref), b.FromEnvelope))
		}
		return out
	}

	n := len(m.Reeds)
	curr := make([]float64, n)
	for i, r := range m.Reeds {
		curr[i] = CurrCents(r, ref)
	}
	out := make([]BeatError, 0, n*(n-1)/2)
	for lo := 0; lo < n; lo++ {
		for hi := lo + 1; hi < n; hi++ {
			out = append(out, row(lo, hi, curr[hi]-curr[lo], false))
		}
	}
	return out
}

// parsePair reads dsp.BeatMeasure.Pair, which is 1-based, into reed indices.
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
