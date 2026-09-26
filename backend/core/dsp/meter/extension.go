package meter

import (
	"math"
	"slices"
)

// ExtractPeaksN generalizes ExtractPeaks to n ranks for registers of more than three reeds:
// n lower-bin seeds, the n strongest in stable scan order, and the same quarter floor. Never used for n <= 3.
func ExtractPeaksN(spectrum []float32, lower, upper, n int) []int {
	top := make([]int, n)
	for k := range top {
		top[k] = lower
	}
	for i := lower; i < upper; i++ {
		v := spectrum[i]
		if !(v > spectrum[i-1] && v >= spectrum[i+1]) {
			continue
		}
		for k := range top {
			if v > spectrum[top[k]] {
				copy(top[k+1:], top[k:n-1])
				top[k] = i
				break
			}
		}
	}
	floor := float64(spectrum[top[0]]) * 0.25
	ranked := make([]int, 0, n)
	for _, index := range top {
		if float64(spectrum[index]) >= floor {
			ranked = append(ranked, index)
		}
	}
	return append(ranked, make([]int, n-len(ranked))...)
}

// AssignPeaksN places the positive ranked bins on n slots: a full set in bin order, otherwise the
// order-keeping placement with the least summed distance to the slot targets (99 for a slot without one).
func AssignPeaksN(ranked []int, targets []float64, referenceBin float64) []int {
	n := len(targets)
	var found []int
	for _, index := range ranked {
		if index > 0 {
			found = append(found, index)
		}
	}
	slices.Sort(found)
	out := make([]int, n)
	if len(found) == n {
		copy(out, found)
		return out
	}
	if len(found) == 0 {
		return out
	}
	distance := func(peak, slot int) float64 {
		if targets[slot] <= Missing {
			return 99
		}
		return math.Abs(targets[slot] - DeviationUnits(float64(found[peak]), referenceBin))
	}
	// cost[i][j]: best total for peaks i.. on slots j.., keeping both in ascending order.
	m := len(found)
	cost := make([][]float64, m+1)
	for i := range cost {
		cost[i] = make([]float64, n+1)
	}
	for i := m - 1; i >= 0; i-- {
		cost[i][n] = math.Inf(1)
		for j := n - 1; j >= 0; j-- {
			cost[i][j] = min(distance(i, j)+cost[i+1][j+1], cost[i][j+1])
		}
	}
	for i, j := 0, 0; i < m; j++ {
		if distance(i, j)+cost[i+1][j+1] <= cost[i][j+1] {
			out[j] = found[i]
			i++
		}
	}
	return out
}
