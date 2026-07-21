package dsp

import (
	"math"
	"sync"
)

// grow returns buf sized to n, reallocating only when it has to.
func grow(buf *[]complex128, n int) []complex128 {
	if cap(*buf) < n {
		*buf = make([]complex128, n)
	}
	*buf = (*buf)[:n]
	return *buf
}

// heterodyne shifts the band at w radians per sample down to baseband. The local oscillator is
// stepped by complex rotation rather than a Sincos per sample (cheaper than the FFT it feeds) and
// re-anchored every reanchorEvery samples so the rotation's rounding cannot accumulate.
func heterodyne(data []float64, out []complex128, w float64) {
	const reanchorEvery = 1024

	var lr, li float64
	dr, di := math.Cos(w), math.Sin(w)

	for i, x := range data {
		if i%reanchorEvery == 0 {
			li, lr = math.Sincos(w * float64(i))
		}
		out[i] = complex(x*lr, x*li)
		lr, li = lr*dr-li*di, lr*di+li*dr
	}
}

const lp2Taps = 63

var (
	lp2Coeffs []float64
	lp2Once   sync.Once
)

// lp2Filter returns the 63-tap windowed-sinc lowpass used before each decimate-by-2, cutoff 0.22 of
// the input rate (headroom below the 0.25 Nyquist of the output rate).
func lp2Filter() []float64 {
	lp2Once.Do(func() {
		const cutoff = 0.22
		h := make([]float64, lp2Taps)
		var sum float64
		for i := range h {
			x := float64(i - lp2Taps/2)
			var s float64
			if x == 0 {
				s = 2 * math.Pi * cutoff
			} else {
				s = math.Sin(2*math.Pi*cutoff*x) / x
			}
			// Blackman window
			w := 0.42 - 0.5*math.Cos(2*math.Pi*float64(i)/(lp2Taps-1)) + 0.08*math.Cos(4*math.Pi*float64(i)/(lp2Taps-1))
			h[i] = s * w
			sum += h[i]
		}
		for i := range h {
			h[i] /= sum
		}
		lp2Coeffs = h
	})
	return lp2Coeffs
}

// lowpassDecimate2 halves the sample rate: windowed-sinc lowpass, keep every second sample, appending
// into out (owned by the caller). The kernel is real and symmetric, and both facts are used because
// this is where the fine stage spends most of its time.
func lowpassDecimate2(in, out []complex128) []complex128 {
	h := lp2Filter()
	const half = lp2Taps / 2

	for i := lp2Taps; i < len(in); i += 2 {
		// The centre tap, which has no partner.
		accRe, accIm := real(in[i-half])*h[half], imag(in[i-half])*h[half]

		for k := 0; k < half; k++ {
			a := in[i-k]
			b := in[i-(lp2Taps-1-k)]
			accRe += (real(a) + real(b)) * h[k]
			accIm += (imag(a) + imag(b)) * h[k]
		}
		out = append(out, complex(accRe, accIm))
	}
	return out
}
