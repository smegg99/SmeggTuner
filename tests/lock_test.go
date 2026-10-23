package tests

import (
	"math"
	"slices"

	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// lockWindow is how many consecutive analyses the reference history holds before it can lock.
const lockWindow = 8

// settleRun is how many analyses into a note the engine may first lock: its fourth native pick, at four
// analyses a native event (dsp settlePicks and nativeEvery).
const settleRun = 3*4 + 1

// lockSpan is the widest deviation span, in units, that the reference history still locks on for a note.
func lockSpan(rate int, a4 float64, note tuning.Note) float64 {
	table := meter.EqualTemperament(a4, 0)
	n := int(note) - 12
	fft := meter.NewWindowPlan(rate, true).FFTSize
	return meter.NewNoteBand(table[n-1], table[n], table[n+1], fft, rate, 1).LockSpanUnits()
}

// steadiness is where the engine's history first accepts note (-1 if never), the narrowest
// first-slot span over lockWindow analyses in a row on it (+Inf if none), and the longest such run.
func steadiness(all []dsp.Measurement, note tuning.Note, limit float64) (at int, span float64, longest int) {
	at, span = -1, math.Inf(1)
	run := 0
	var window []float64
	for i, m := range all {
		if m.Frame == dsp.FrameNone {
			continue
		}
		fresh := m.Frame == dsp.FrameMeasured || m.Frame == dsp.FrameIneligible
		if !fresh || m.Note != note || !m.Reeds[0].Present {
			window, run = window[:0], 0
			continue
		}
		run++
		longest = max(longest, run)
		window = append(window, m.Reeds[0].Dev)
		if len(window) > lockWindow {
			window = window[1:]
		}
		if len(window) == lockWindow {
			span = min(span, slices.Max(window)-slices.Min(window))
			if at < 0 && span <= limit && run >= settleRun {
				at = i
			}
		}
	}
	return at, span, longest
}
