package dsp

import (
	"math"
	"math/cmplx"
)

// projAt projects the band onto one line at f: sum w^2 * z * exp(-i2pi f t).
func (pb *pairBand) projAt(f float64) complex128 {
	var acc complex128
	d := cmplx.Exp(complex(0, -2*math.Pi*(f-pb.center)/pb.rate))
	rot := complex(1, 0)
	for i, z := range pb.band {
		acc += complex(pb.w2[i], 0) * z * rot
		rot *= d
	}
	return acc
}

// gramAt is the inner product of two lines d apart: sum w^2 * exp(i2pi d t).
func (pb *pairBand) gramAt(d float64) complex128 {
	var acc complex128
	st := cmplx.Exp(complex(0, 2*math.Pi*d/pb.rate))
	rot := complex(1, 0)
	for _, w := range pb.w2 {
		acc += complex(w, 0) * rot
		rot *= st
	}
	return acc
}

// proj and gram read the tables; exact says read the band instead, which the final solve does.
func (pb *pairBand) proj(f float64, exact bool) complex128 {
	if exact {
		return pb.projAt(f)
	}
	return interpTab(pb.projTab, pb.projLo, pb.projStep, f)
}

func (pb *pairBand) gram(d float64, exact bool) complex128 {
	if d < 0 {
		return cmplx.Conj(pb.gram(-d, exact))
	}
	if exact {
		return pb.gramAt(d)
	}
	return interpTab(pb.gramTab, pb.gramLo, pb.gramStep, d)
}

// interpTab is a Catmull-Rom cubic through the four entries around x. Off the ends it returns zero:
// two lines that far apart do not see each other.
func interpTab(tab []complex128, lo, step float64, x float64) complex128 {
	if len(tab) < 4 {
		return 0
	}
	u := (x - lo) / step
	i := int(math.Floor(u))
	if i < 1 || i > len(tab)-3 {
		return 0
	}
	t := u - float64(i)
	p0, p1, p2, p3 := tab[i-1], tab[i], tab[i+1], tab[i+2]
	c0 := p1
	c1 := 0.5 * (p2 - p0)
	c2 := p0 - 2.5*p1 + 2*p2 - 0.5*p3
	c3 := 0.5*(p3-p0) + 1.5*(p1-p2)
	tc := complex(t, 0)
	return c0 + tc*(c1+tc*(c2+tc*c3))
}

// gramSet is every value of the window's transform that one placement of the model can ask for,
// gathered once. The model's lines sit at (k-j)*b offset by 0, +-fa or +-2fa, and (k-j) runs -2..2, so
// there are only ever twenty-five distinct numbers underneath.
type gramSet struct {
	// [m+2] holds the transform at m*b + shift, for m in -2..2.
	at, up, dn, up2, dn2 [5]complex128
}

func (pb *pairBand) gramSet(b, fa float64, exact bool) gramSet {
	var g gramSet
	for m := -2; m <= 2; m++ {
		d := float64(m) * b
		g.at[m+2] = pb.gram(d, exact)
		if fa <= 0 {
			continue
		}
		g.up[m+2] = pb.gram(d+fa, exact)
		g.dn[m+2] = pb.gram(d-fa, exact)
		g.up2[m+2] = pb.gram(d+2*fa, exact)
		g.dn2[m+2] = pb.gram(d-2*fa, exact)
	}
	return g
}

// combGram is the Gram matrix of the comb's lines under the bellows: entry (j,k) is the inner product
// of line k with line j, each swung by its own stroke. u_j and u_k are real, so their product is five
// lines at 0, +-fa and +-2fa, read off the window's transform. Hermitian by construction.
func (pb *pairBand) combGram(g gramSet, fa float64, c [combLines]complex128) ([combLines][combLines]complex128, bool) {
	var m [combLines][combLines]complex128
	on := fa > 0
	// Hermitian, so only the upper triangle is built and the rest is its mirror.
	for j := 0; j < combLines; j++ {
		for k := j; k < combLines; k++ {
			d := k - j + 2
			if !on {
				m[j][k] = g.at[d]
			} else {
				cj, ck := c[j], c[k]
				m[j][k] = complex(1+2*real(ck*cmplx.Conj(cj)), 0)*g.at[d] +
					(ck+cj)*g.up[d] +
					cmplx.Conj(ck+cj)*g.dn[d] +
					(ck*cj)*g.up2[d] +
					cmplx.Conj(ck*cj)*g.dn2[d]
			}
			if k != j {
				m[k][j] = cmplx.Conj(m[j][k])
			}
		}
	}
	if abs2(m[0][0]) <= 0 {
		return m, false
	}
	return m, true
}
