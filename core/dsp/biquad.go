package dsp

import "math"

// Biquad is a cascade of direct-form-I second-order IIR sections using RBJ cookbook coefficients, normalized so a0 == 1.
type Biquad struct {
	secs   []biquadSection
	primed bool
}

type biquadSection struct {
	b0, b1, b2, a1, a2 float64
	x1, x2, y1, y2     float64
}

func notchSection(sampleRate, freq, q float64) biquadSection {
	w0 := 2 * math.Pi * freq / sampleRate
	alpha := math.Sin(w0) / (2 * q)
	b0, b1, b2 := 1.0, -2*math.Cos(w0), 1.0
	a0, a1, a2 := 1+alpha, -2*math.Cos(w0), 1-alpha
	return biquadSection{b0: b0 / a0, b1: b1 / a0, b2: b2 / a0, a1: a1 / a0, a2: a2 / a0}
}

func highpassSection(sampleRate, cutoff float64) biquadSection {
	w0 := 2 * math.Pi * cutoff / sampleRate
	q := math.Sqrt2 / 2
	alpha := math.Sin(w0) / (2 * q)
	cosw := math.Cos(w0)
	b0 := (1 + cosw) / 2
	b1 := -(1 + cosw)
	b2 := (1 + cosw) / 2
	a0 := 1 + alpha
	a1 := -2 * cosw
	a2 := 1 - alpha
	return biquadSection{b0: b0 / a0, b1: b1 / a0, b2: b2 / a0, a1: a1 / a0, a2: a2 / a0}
}

// NewNotch builds a notch at freq; q around 30 is narrow enough for mains hum without touching nearby musical content.
func NewNotch(sampleRate, freq, q float64) *Biquad {
	return &Biquad{secs: []biquadSection{notchSection(sampleRate, freq, q)}}
}

// NewHighpass builds a rumble filter from two cascaded 2nd-order Butterworth sections (Linkwitz-Riley, 24 dB/oct).
func NewHighpass(sampleRate, cutoff float64) *Biquad {
	return &Biquad{secs: []biquadSection{
		highpassSection(sampleRate, cutoff),
		highpassSection(sampleRate, cutoff),
	}}
}

func (b *Biquad) step(x float64) float64 {
	for i := range b.secs {
		s := &b.secs[i]
		y := s.b0*x + s.b1*s.x1 + s.b2*s.x2 - s.a1*s.y1 - s.a2*s.y2
		s.x2, s.x1 = s.x1, x
		s.y2, s.y1 = s.y1, y
		x = y
	}
	return x
}

// Process filters samples in place, carrying state across calls.
func (b *Biquad) Process(samples []float32) {
	if !b.primed && len(samples) > 0 {
		b.primed = true
		// Settle state on an odd reflection of the first chunk (filtfilt-style padding) so narrow
		// filters do not ring for thousands of samples after startup.
		x0 := float64(samples[0])
		for i := len(samples) - 1; i >= 1; i-- {
			b.step(2*x0 - float64(samples[i]))
		}
	}
	for i, s := range samples {
		samples[i] = float32(b.step(float64(s)))
	}
}
