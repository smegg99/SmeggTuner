package dsp

import "math"

// PairDepth is how deeply two reeds of amplitude ratio r MUST swing the amplitude between them: the
// first Fourier coefficient over the mean of A*sqrt(1 + r^2 + 2r*cos(2*pi*b*t)). It rises with r to
// 2/3 at r = 1 and cannot pass it. It is what PairFit.Depth is weighed against: two lines the right
// distance apart can be faked, two lines that interfere cannot.
func PairDepth(r float64) float64 {
	if r <= 0 {
		return 0
	}
	if r > 1 {
		r = 1 / r
	}
	// Smooth except as r approaches 1, where it cusps at the beat's null; a few hundred points carry that.
	const steps = 720
	var mean, first float64
	for i := 0; i < steps; i++ {
		th := 2 * math.Pi * float64(i) / steps
		e := math.Sqrt(1 + r*r + 2*r*math.Cos(th))
		mean += e
		first += e * math.Cos(th)
	}
	if mean <= 0 {
		return 0
	}
	return 2 * first / mean
}
