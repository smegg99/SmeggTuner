package dsp

import (
	"math"
	"testing"
)

// The trace keeps the signal's level: a quiet block draws small, it is not stretched to full height.
func TestDecimateWaveKeepsLevel(t *testing.T) {
	for _, amp := range []float32{0.02, 0.25, 0.9} {
		samples := make([]float32, 4800)
		for i := range samples {
			samples[i] = amp * float32(math.Sin(2*math.Pi*440*float64(i)/48000))
		}
		var e Engine
		e.decimateWave(samples)

		var peak float32
		for _, v := range e.wave {
			peak = max(peak, abs32(v))
		}
		if math.Abs(float64(peak-amp)) > float64(amp)*0.01 {
			t.Errorf("amplitude %v: trace peak %v, want the signal's own level", amp, peak)
		}
	}
}
