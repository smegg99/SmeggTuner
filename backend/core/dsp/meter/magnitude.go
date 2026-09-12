package meter

import "math"

// MagnitudeBounds is the computed bin interval [Lower, Upper); every other magnitude is zero.
type MagnitudeBounds struct {
	Lower, Upper int
}

// NewMagnitudeBounds widens table entries 3 and 109 by 3 percent; ok needs a positive finite clock and 1 <= Lower <= Upper.
func NewMagnitudeBounds(fftSize, sampleRate int, rateFactor float64, table *NoteTable) (MagnitudeBounds, bool) {
	transformCheckSize(fftSize)
	clock := float64(sampleRate) * rateFactor
	if !(clock > 0) || math.IsInf(clock, 1) {
		return MagnitudeBounds{}, false
	}
	half := fftSize / 2
	lo := float64(table[3] * float64(half))
	hi := float64(table[109] * float64(half))
	a := float64(((lo+lo)/clock)*0.97) + 0.5
	b := float64(((hi+hi)/clock)*1.03) + 0.5
	if !(a >= -(1<<31) && a < 1<<31 && b >= -(1<<31) && b < 1<<31) {
		return MagnitudeBounds{}, false
	}
	m := MagnitudeBounds{Lower: max(0, int(a)), Upper: min(half, int(b))}
	return m, 1 <= m.Lower && m.Lower <= m.Upper
}

// Magnitudes writes and returns len(packed)/2 binary32 magnitudes sqrt(r*r+s*s)/M, zero outside b; dst may alias packed.
func Magnitudes(dst, packed []float32, b MagnitudeBounds) []float32 {
	half := len(packed) / 2
	dst = dst[:half]
	divisor := float32(half)
	for k := b.Lower; k < b.Upper; k++ {
		r, s := packed[2*k], packed[2*k+1]
		dst[k] = float32(math.Sqrt(float64(float32(r*r)+float32(s*s)))) / divisor
	}
	clear(dst[:b.Lower])
	clear(dst[b.Upper:])
	return dst
}
