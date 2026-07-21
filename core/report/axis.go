package report

import "math"

func padRange(minC, maxC float64) (float64, float64) {
	span := maxC - minC
	if span < minSpan {
		mid := (minC + maxC) / 2
		minC, maxC = mid-minSpan/2, mid+minSpan/2
		span = minSpan
	}
	pad := span * 0.08
	return minC - pad, maxC + pad
}

func yTicks(minC, maxC float64) []float64 {
	step := niceStep((maxC - minC) / 5)
	var out []float64
	for v := math.Ceil(minC/step) * step; v <= maxC; v += step {
		out = append(out, v)
	}
	return out
}

func niceStep(raw float64) float64 {
	if raw <= 0 {
		return 1
	}
	mag := math.Pow(10, math.Floor(math.Log10(raw)))
	switch n := raw / mag; {
	case n <= 1:
		return mag
	case n <= 2:
		return 2 * mag
	case n <= 5:
		return 5 * mag
	default:
		return 10 * mag
	}
}
