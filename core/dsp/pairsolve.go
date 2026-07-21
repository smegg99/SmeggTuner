package dsp

import (
	"math"
	"math/cmplx"
)

// amSolution is one placement of the model: a comb of combLines lines spaced b apart about f, under a
// bellows of depth 2|c| stroking at fa.
type amSolution struct {
	f, b, fa  float64
	c         [combLines]complex128
	a         [combLines]complex128
	explained float64
}

// amRounds is how many times the comb and the bellows are solved against each other. It converges from
// below: the first round sees only the sidebands the plain comb missed, and each round recovers a
// share of the rest. Measured, the stroke settles by six rounds; the beat follows it down.
const amRounds = 6

// solve fits the model at one placement: the comb amplitudes a and the bellows c, by alternation,
// starting from no bellows. The two are not linear in each other but each is linear given the other,
// and both solves are closed form. fa of zero means the plain three-line comb.
func (pb *pairBand) solve(f, b, fa float64, exact bool) (amSolution, bool) {
	return pb.solveRounds(f, b, fa, exact, amRounds)
}

// solveRounds is solve with an explicit budget of alternations. The stroke-rate sweep uses a short one:
// it is only ranking rates, and a rate that will win on six rounds wins on two.
func (pb *pairBand) solveRounds(f, b, fa float64, exact bool, rounds int) (amSolution, bool) {
	s := amSolution{f: f, b: b, fa: fa}
	fk := [combLines]float64{f - b, f, f + b}

	// p is the band projected onto the comb's own lines, and it never changes: the bellows moves what
	// the MODEL's lines are, not where the band's energy sits.
	var p [combLines]complex128
	for k := 0; k < combLines; k++ {
		p[k] = pb.proj(fk[k], exact)
	}
	gs := pb.gramSet(b, fa, exact)
	var noAM [combLines]complex128
	g, ok := pb.combGram(gs, 0, noAM)
	if !ok {
		return s, false
	}
	a, ok := solveHermitian(g, p)
	if !ok {
		return s, false
	}
	s.a, s.explained = a, explainedBy(a, p, pb.energy)

	if fa <= 0 {
		return s, true
	}

	// Where the bellows would put its sidebands, and how much of the band is there.
	var sbUp, sbDn [combLines]complex128
	for k := 0; k < combLines; k++ {
		sbUp[k] = pb.proj(fk[k]+fa, exact)
		sbDn[k] = pb.proj(fk[k]-fa, exact)
	}
	for round := 0; round < rounds; round++ {
		c, _, ok := pb.solveAM(gs, s.a, sbUp, sbDn)
		if !ok {
			break
		}
		g, ok := pb.combGram(gs, fa, c)
		if !ok {
			break
		}
		// The band projected onto the comb's lines AS THE BELLOWS SWINGS THEM: each real sideband folds
		// the band back onto the line it hangs off.
		var pu [combLines]complex128
		for k := 0; k < combLines; k++ {
			pu[k] = p[k] + c[k]*sbDn[k] + cmplx.Conj(c[k])*sbUp[k]
		}
		na, ok := solveHermitian(g, pu)
		if !ok {
			break
		}
		e := explainedBy(na, pu, pb.energy)
		if e <= s.explained {
			break // the bellows bought nothing more; keep what stood without it
		}
		s.a, s.c, s.explained = na, c, e
	}
	return s, true
}

// explainedBy is the fraction of the band's energy a least-squares fit accounts for: a^H p. It cannot
// exceed one, and a fit that says it does has a matrix and projections describing different models.
func explainedBy(a, p [combLines]complex128, energy float64) float64 {
	var e float64
	for k := 0; k < combLines; k++ {
		e += real(cmplx.Conj(a[k]) * p[k])
	}
	return e / energy
}

// solveAM finds the bellows, given the comb it swings. The residual is fitted with each comb line
// shifted up by fa and scaled by c_k, and shifted down and scaled by the CONJUGATE of the same c_k -
// that conjugate tie is the whole content of "this is a bellows, not another reed", and makes the
// problem real: two unknowns per tooth, six in all, closed form.
//
// One stroke rate is shared (the constraint doing the work), but a coefficient per reed, because each
// reed answers the stroke through its own tongue with its own lag; forcing one coefficient leaks the
// difference into the spacing. No bellows may claim a sideband as loud as reedPeakFloor of the line it
// hangs off, because a line that loud is a reed everywhere else, so each is capped there.
func (pb *pairBand) solveAM(g gramSet, a, sbUp, sbDn [combLines]complex128) (c [combLines]complex128, gain float64, ok bool) {
	// The basis: for each tooth k, A_k is that tooth shifted up and B_k shifted down, both carrying the
	// tooth's amplitude. c_k = x_k + i*y_k enters as x_k*(A_k + B_k) + y_k*i*(A_k - B_k).
	var aa, ab [combLines][combLines]complex128
	for k := 0; k < combLines; k++ {
		for l := 0; l < combLines; l++ {
			coef := a[k] * cmplx.Conj(a[l])
			d := k - l + 2
			aa[k][l] = coef * g.at[d]
			ab[k][l] = coef * g.up2[d]
		}
	}
	// What the band has where the bellows would put something, less what the comb accounts for there.
	var rA, rB [combLines]complex128
	for k := 0; k < combLines; k++ {
		var pA, pB complex128
		for j := 0; j < combLines; j++ {
			d := j - k + 2
			pA += a[j] * g.dn[d]
			pB += a[j] * g.up[d]
		}
		rA[k] = cmplx.Conj(a[k]) * (sbUp[k] - pA)
		rB[k] = cmplx.Conj(a[k]) * (sbDn[k] - pB)
	}

	var m [6][6]float64
	var v [6]float64
	for k := 0; k < combLines; k++ {
		for l := 0; l < combLines; l++ {
			akal, bkbl := aa[k][l], aa[k][l]
			akbl, bkal := ab[k][l], cmplx.Conj(ab[l][k])
			// U = A + B, V = i(A - B).
			uu := akal + akbl + bkal + bkbl
			uv := complex(0, -1) * (akal - akbl + bkal - bkbl)
			vu := complex(0, 1) * (akal + akbl - bkal - bkbl)
			vv := akal - akbl - bkal + bkbl
			m[2*k][2*l] = real(uu)
			m[2*k][2*l+1] = real(uv)
			m[2*k+1][2*l] = real(vu)
			m[2*k+1][2*l+1] = real(vv)
		}
		v[2*k] = real(rA[k] + rB[k])
		v[2*k+1] = real(complex(0, -1) * (rA[k] - rB[k]))
	}

	x, ok := solveSym6(m, v)
	if !ok {
		return c, 0, false
	}
	// x^T v is the energy the bellows takes out of what the comb could not explain; findAM ranks stroke rates by it.
	for p := 0; p < 6; p++ {
		gain += x[p] * v[p]
	}
	for k := 0; k < combLines; k++ {
		ck := complex(x[2*k], x[2*k+1])
		if m2 := abs2(ck); m2 > reedPeakFloor*reedPeakFloor {
			ck *= complex(reedPeakFloor/math.Sqrt(m2), 0)
		}
		c[k] = ck
	}
	return c, gain, true
}
