package meter

import "math"

// transformTrig is pocketfft's two-level table of exp(2*pi*i*k/n): v1[k&mask]*v2[k>>shift].
type transformTrig struct {
	v1, v2      [][2]float64
	mask, shift int
}

func newTransformTrig(n int) transformTrig {
	ang := math.Pi / float64(4*n)
	nval := (n + 2) / 2
	shift := 1
	for (1<<shift)*(1<<shift) < nval {
		shift++
	}
	mask := 1<<shift - 1
	t := transformTrig{v1: make([][2]float64, mask+1), v2: make([][2]float64, (nval+mask)/(mask+1)), mask: mask, shift: shift}
	t.v1[0], t.v2[0] = [2]float64{1, 0}, [2]float64{1, 0}
	for i := 1; i < len(t.v1); i++ {
		t.v1[i] = transformOctant(i, n, ang)
	}
	for i := 1; i < len(t.v2); i++ {
		t.v2[i] = transformOctant(i*(mask+1), n, ang)
	}
	return t
}

// at is valid for 2*k <= n, the only range a real transform reads.
func (t transformTrig) at(k int) (float64, float64) {
	a, b := t.v1[k&t.mask], t.v2[k>>t.shift]
	return float64(a[0]*b[0]) - float64(a[1]*b[1]), float64(a[0]*b[1]) + float64(a[1]*b[0])
}

// transformOctant folds 2*pi*x/n into the first octant before evaluating sine and cosine.
func transformOctant(x, n int, ang float64) [2]float64 {
	x <<= 3
	flip := x >= 4*n
	if flip {
		x = 8*n - x
	}
	var r [2]float64
	switch {
	case x < n:
		r[1], r[0] = transformSinCos(float64(x) * ang)
	case x < 2*n:
		r[0], r[1] = transformSinCos(float64(2*n-x) * ang)
	case x < 3*n:
		r[0], r[1] = transformSinCos(float64(x-2*n) * ang)
		r[0] = -r[0]
	default:
		r[1], r[0] = transformSinCos(float64(4*n-x) * ang)
		r[0] = -r[0]
	}
	if flip {
		r[1] = -r[1]
	}
	return r
}

// transformSinCos rounds a double-double Taylor series once; it is correctly rounded on every twiddle argument up to 2^21.
func transformSinCos(a float64) (sin, cos float64) {
	a2h := a * a
	a2l := math.FMA(a, a, -a2h)
	sh, sl, ch, cl := 1.0, 0.0, 1.0, 0.0
	for k := 15; k >= 1; k-- {
		sh, sl = transformSeriesStep(sh, sl, a2h, a2l, float64(2*k*(2*k+1)))
		ch, cl = transformSeriesStep(ch, cl, a2h, a2l, float64((2*k-1)*2*k))
	}
	p := float64(a * sh)
	return p + (math.FMA(a, sh, -p) + float64(a*sl)), ch + cl
}

// transformSeriesStep returns 1 - (h+l)*(a2h+a2l)/d in double-double.
func transformSeriesStep(h, l, a2h, a2l, d float64) (float64, float64) {
	ph := float64(h * a2h)
	pl := math.FMA(h, a2h, -ph) + (float64(h*a2l) + float64(l*a2h))
	ph, pl = ph+pl, pl-((ph+pl)-ph)
	qh := ph / d
	ql := (math.FMA(-qh, d, ph) + pl) / d
	s := 1 - qh
	v := s - 1
	e := (1 - (s - v)) + (-qh - v) - ql
	return s + e, e - ((s + e) - s)
}
