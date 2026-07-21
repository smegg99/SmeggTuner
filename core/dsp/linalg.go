package dsp

import (
	"math"
)

// solve3 is Gaussian elimination with partial pivoting on a real 3x3.
func solve3(m [3][3]float64, v [3]float64) ([3]float64, bool) {
	var x [3]float64
	scale := math.Abs(m[0][0])
	if scale <= 0 {
		return x, false
	}
	for col := 0; col < 3; col++ {
		p := col
		for r := col + 1; r < 3; r++ {
			if math.Abs(m[r][col]) > math.Abs(m[p][col]) {
				p = r
			}
		}
		if math.Abs(m[p][col]) < 1e-12*scale {
			return x, false
		}
		m[col], m[p] = m[p], m[col]
		v[col], v[p] = v[p], v[col]
		for r := col + 1; r < 3; r++ {
			f := m[r][col] / m[col][col]
			for c := col; c < 3; c++ {
				m[r][c] -= f * m[col][c]
			}
			v[r] -= f * v[col]
		}
	}
	for r := 2; r >= 0; r-- {
		s := v[r]
		for c := r + 1; c < 3; c++ {
			s -= m[r][c] * x[c]
		}
		x[r] = s / m[r][r]
	}
	return x, true
}

// abs2 is the squared magnitude, for the many places that only need to compare two of them.
func abs2(c complex128) float64 { return real(c)*real(c) + imag(c)*imag(c) }

// solveSym6 is Gaussian elimination with partial pivoting on a real 6x6. It duplicates solve3 at a
// different size on purpose: Go has no const generics, and writing these over slices allocates inside
// the search loop where the whole fit holds to 22 allocations.
func solveSym6(m [6][6]float64, v [6]float64) ([6]float64, bool) {
	var x [6]float64
	var scale float64
	for i := 0; i < 6; i++ {
		if a := math.Abs(m[i][i]); a > scale {
			scale = a
		}
	}
	if scale <= 0 {
		return x, false
	}
	for col := 0; col < 6; col++ {
		p := col
		for r := col + 1; r < 6; r++ {
			if math.Abs(m[r][col]) > math.Abs(m[p][col]) {
				p = r
			}
		}
		if math.Abs(m[p][col]) < 1e-10*scale {
			return x, false
		}
		m[col], m[p] = m[p], m[col]
		v[col], v[p] = v[p], v[col]
		for r := col + 1; r < 6; r++ {
			f := m[r][col] / m[col][col]
			for c := col; c < 6; c++ {
				m[r][c] -= f * m[col][c]
			}
			v[r] -= f * v[col]
		}
	}
	for r := 5; r >= 0; r-- {
		s := v[r]
		for c := r + 1; c < 6; c++ {
			s -= m[r][c] * x[c]
		}
		x[r] = s / m[r][r]
	}
	return x, true
}

// solveHermitian solves the comb's normal equations by Gaussian elimination with partial pivoting.
// ok is false on a singular matrix, meaning two comb lines sat too close to be told apart.
func solveHermitian(m [combLines][combLines]complex128, b [combLines]complex128) ([combLines]complex128, bool) {
	var x [combLines]complex128
	// Squared magnitudes throughout: only which of two is larger matters, and a sqrt in the innermost
	// loop was six percent of the whole fit.
	scale := abs2(m[0][0])
	if scale <= 0 {
		return x, false
	}
	for col := 0; col < combLines; col++ {
		p := col
		for r := col + 1; r < combLines; r++ {
			if abs2(m[r][col]) > abs2(m[p][col]) {
				p = r
			}
		}
		if abs2(m[p][col]) < 1e-24*scale {
			return x, false
		}
		m[col], m[p] = m[p], m[col]
		b[col], b[p] = b[p], b[col]
		for r := col + 1; r < combLines; r++ {
			f := m[r][col] / m[col][col]
			for c := col; c < combLines; c++ {
				m[r][c] -= f * m[col][c]
			}
			b[r] -= f * b[col]
		}
	}
	for r := combLines - 1; r >= 0; r-- {
		s := b[r]
		for c := r + 1; c < combLines; c++ {
			s -= m[r][c] * x[c]
		}
		x[r] = s / m[r][r]
	}
	return x, true
}

// goldenMax finds the maximum of f on [lo, hi] by golden section, to well inside a thousandth of a cent.
func goldenMax(f func(float64) float64, lo, hi float64) float64 {
	const invPhi = 0.6180339887498949
	const iters = 24
	if hi <= lo {
		return lo
	}
	a, b := lo, hi
	c := b - invPhi*(b-a)
	d := a + invPhi*(b-a)
	fc, fd := f(c), f(d)
	for i := 0; i < iters; i++ {
		if fc > fd {
			b, d, fd = d, c, fc
			c = b - invPhi*(b-a)
			fc = f(c)
		} else {
			a, c, fc = c, d, fd
			d = a + invPhi*(b-a)
			fd = f(d)
		}
	}
	return (a + b) / 2
}
