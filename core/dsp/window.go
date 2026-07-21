package dsp

import (
	"math"
	"sync"
)

var hannCache sync.Map // int -> []float64

// Hann returns a Hann window of length n, cached per n. The returned slice must not be mutated.
func Hann(n int) []float64 {
	if w, ok := hannCache.Load(n); ok {
		return w.([]float64)
	}
	w := make([]float64, n)
	for i := range w {
		w[i] = 0.5 * (1 - math.Cos(2*math.Pi*float64(i)/float64(n-1)))
	}
	hannCache.Store(n, w)
	return w
}
