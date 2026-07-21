package dsp

import (
	"math"
)

// bandTaps is the filter length: odd for a whole-sample delay, long enough that the transition band
// clears a reed's harmonics at every note a pair can merge at.
const bandTaps = 63

// bandLimit low-passes the zoom's baseband to +-cutHz around the note.
//
// Zoom's TimeSeries is not the analysis band, only what survived the decimation chain, so at low
// notes the reed's own second harmonic is still in it: it fills a beat's nulls (reading the swing as
// 0.54 where the pair swings 0.65) and holds energy no model of the note can explain. So the comb is
// fitted against the note's own neighbourhood, taken here. The filter's settling length is dropped
// off the front, which does not matter to lines fitted at known frequencies.
func bandLimit(z []complex128, rate, cutHz float64) []complex128 {
	if rate <= 0 || cutHz <= 0 || len(z) <= bandTaps {
		return z
	}
	fc := cutHz / rate
	if fc >= 0.5 {
		return z
	}
	h := make([]float64, bandTaps)
	var sum float64
	for i := range h {
		x := float64(i - bandTaps/2)
		s := 2 * fc
		if x != 0 {
			s = math.Sin(2*math.Pi*fc*x) / (math.Pi * x)
		}
		w := 0.42 - 0.5*math.Cos(2*math.Pi*float64(i)/(bandTaps-1)) +
			0.08*math.Cos(4*math.Pi*float64(i)/(bandTaps-1))
		h[i] = s * w
		sum += h[i]
	}
	for i := range h {
		h[i] /= sum
	}

	out := make([]complex128, 0, len(z)-bandTaps+1)
	for i := bandTaps - 1; i < len(z); i++ {
		var acc complex128
		for k := 0; k < bandTaps; k++ {
			acc += z[i-k] * complex(h[k], 0)
		}
		out = append(out, acc)
	}
	return out
}
