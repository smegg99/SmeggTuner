package meter

import (
	"math"
	"slices"
)

// ExtractPeaks ranks local maxima in [lower, upper) with three lower-bin seeds first; ranks under a quarter of the top are 0.
func ExtractPeaks(spectrum []float32, lower, upper int) [3]int {
	top := [3]int{lower, lower, lower}
	for i := lower; i < upper; i++ {
		v := spectrum[i]
		if !(v > spectrum[i-1] && v >= spectrum[i+1]) {
			continue
		}
		// The three leaders of a stable descending sort: a tie stays behind the earlier candidate.
		for k := range top {
			if v > spectrum[top[k]] {
				copy(top[k+1:], top[k:2])
				top[k] = i
				break
			}
		}
	}
	floor := float64(spectrum[top[0]]) * 0.25
	var ranked [3]int
	n := 0
	for _, index := range top {
		if float64(spectrum[index]) >= floor {
			ranked[n] = index
			n++
		}
	}
	return ranked
}

// AssignPeaks maps ranked bins onto reed slots 1..3; selected lists the measured slots in increasing order, targets their deviations.
func AssignPeaks(ranked [3]int, selected []int, targets [3]float64, referenceBin float64) [3]int {
	var found [3]int
	n := 0
	for _, index := range ranked[:len(selected)] {
		if index > 0 {
			found[n] = index
			n++
		}
	}
	slices.Sort(found[:n])
	var out [3]int
	if n == 0 {
		return out
	}
	if n == len(selected) {
		for i, slot := range selected {
			out[slot-1] = found[i]
		}
		return out
	}
	low := DeviationUnits(float64(found[0]), referenceBin)
	var distance [3]float64
	for slot := 1; slot <= 3; slot++ {
		distance[slot-1] = 99
		if slices.Contains(selected, slot) && targets[slot-1] > Missing {
			distance[slot-1] = math.Abs(targets[slot-1] - low)
		}
	}
	if n == 1 {
		winner := 0
		if distance[1] < distance[0] && distance[1] < distance[2] {
			winner = 1
		} else if distance[2] < distance[0] && distance[2] < distance[1] {
			winner = 2
		}
		out[winner] = found[0]
		return out
	}
	lo, hi := found[0], found[1]
	if distance[1] < distance[0] {
		return [3]int{0, lo, hi}
	}
	high := DeviationUnits(float64(hi), referenceBin)
	if math.Abs(targets[2]-high) < math.Abs(targets[1]-high) {
		return [3]int{lo, 0, hi}
	}
	return [3]int{lo, hi, 0}
}

// Refined is one value per slot; a missing slot has bin 0, offset 0 and deviation Missing.
type Refined struct {
	Bins       []int
	Offsets    []float64
	Deviations []float64
}

// DisplayEligible is the ordinary early-return gate: every present slot within +/-58.
func (r Refined) DisplayEligible() bool {
	for _, d := range r.Deviations {
		if d != Missing && math.Abs(d) > 58 {
			return false
		}
	}
	return true
}

func RefineSlots(spectrum []float32, assigned []int, referenceBin float64) Refined {
	r := Refined{Bins: append([]int(nil), assigned...), Offsets: make([]float64, len(assigned)), Deviations: make([]float64, len(assigned))}
	for slot, index := range assigned {
		if index == 0 {
			r.Deviations[slot] = Missing
			continue
		}
		r.Offsets[slot] = Interpolate(spectrum[index-1], spectrum[index], spectrum[index+1])
		r.Deviations[slot] = DeviationUnits(float64(index)+r.Offsets[slot], referenceBin)
	}
	return r
}
