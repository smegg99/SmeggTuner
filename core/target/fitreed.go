package target

import (
	"cmp"
	"math"
	"slices"

	"smegg.me/smeggtuner/core/tuning"
)

// fitRounds is how many times the fit is redone after throwing out what the last
// fit exposed. Two converges; the third is a safety net.
const fitRounds = 3

// minBinReadings is the fewest readings an octave may anchor on its own; a thinner
// octave the curve would just pass through, so it is merged into a neighbour.
const minBinReadings = 3

// smoothing is how far an anchor is pulled towards the straight line between its
// neighbours: half, which damps the jitter without bending a straight trend.
const smoothing = 0.5

// point is one reed of one reading: what it read, and which octave it fell in.
type point struct {
	note tuning.Note
	bin  int
	curr float64
}

// knot is one reed's fitted value at one anchor note.
type knot struct {
	note tuning.Note
	val  float64
}

// fitReed fits one reed, returning its knots and which points to reject: fit, throw
// out what it exposes, fit again.
func fitReed(pts []point, limit float64) ([]knot, []bool) {
	mergeThinBins(pts)

	bad := make([]bool, len(pts))
	ks := medianKnots(pts, bad)
	for round := 1; round < fitRounds; round++ {
		next := mark(pts, ks, limit)
		if same(next, bad) {
			break
		}
		bad = next
		ks = medianKnots(pts, bad)
	}
	return ks, bad
}

// medianKnots takes each octave's median note and value together, so the anchor
// lands on a pair that actually occurred, then pulls it halfway towards the line
// between its neighbours. An octave with every reading thrown out gets no anchor.
func medianKnots(pts []point, bad []bool) []knot {
	byBin := map[int][]point{}
	for i, p := range pts {
		if !bad[i] {
			byBin[p.bin] = append(byBin[p.bin], p)
		}
	}
	ks := make([]knot, 0, len(byBin))
	for _, in := range byBin {
		notes := make([]tuning.Note, len(in))
		vals := make([]float64, len(in))
		for i, p := range in {
			notes[i] = p.note
			vals[i] = p.curr
		}
		slices.Sort(notes)
		slices.Sort(vals)
		// Lower of the two middles, not their average, so on a rising trend the
		// anchor is a note and a value that occurred together.
		mid := (len(in) - 1) / 2
		ks = append(ks, knot{note: notes[mid], val: vals[mid]})
	}
	slices.SortFunc(ks, func(a, b knot) int { return cmp.Compare(a.note, b.note) })

	if len(ks) < 3 {
		return ks
	}
	src := append([]knot(nil), ks...)
	for i := 1; i < len(ks)-1; i++ {
		a, b := src[i-1], src[i+1]
		t := float64(src[i].note-a.note) / float64(b.note-a.note)
		line := a.val + t*(b.val-a.val)
		ks[i].val = (1-smoothing)*src[i].val + smoothing*line
	}
	return ks
}

// mergeThinBins folds an octave with too few readings into its neighbour, once,
// before any rejection, so the bins do not shift underneath the rounds.
func mergeThinBins(pts []point) {
	count := map[int]int{}
	for _, p := range pts {
		count[p.bin]++
	}
	for len(count) > 1 {
		bins := make([]int, 0, len(count))
		for b := range count {
			bins = append(bins, b)
		}
		slices.Sort(bins)

		thin := -1
		for i, b := range bins {
			if count[b] < minBinReadings {
				thin = i
				break
			}
		}
		if thin < 0 {
			return
		}
		into := 1 // a bottom octave's only neighbour is the one above it
		if thin > 0 {
			into = thin - 1
		}
		from, to := bins[thin], bins[into]
		for i := range pts {
			if pts[i].bin == from {
				pts[i].bin = to
			}
		}
		count[to] += count[from]
		delete(count, from)
	}
}

// mark flags the points further from the trend than limit. A fixed window, not
// one that widens with the scatter: a rank where a third of the reeds drifted
// the same way would otherwise open the window until the drift reads as tremolo.
func mark(pts []point, ks []knot, limit float64) []bool {
	out := make([]bool, len(pts))
	for i, p := range pts {
		out[i] = math.Abs(p.curr-evalKnots(ks, p.note)) > limit
	}
	return out
}

func same(a, b []bool) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// anchorNotes is every note any reed was fitted at, plus the ends of the
// recorded range. The union (reeds are fitted apart, on possibly different
// notes) reproduces each reed's own fit exactly; the ends keep the outermost
// half-octaves from falling outside the anchors and being held flat.
func anchorNotes(fits [][]knot, lo, hi tuning.Note) []tuning.Note {
	ns := []tuning.Note{lo, hi}
	for _, ks := range fits {
		for _, k := range ks {
			ns = append(ns, k.note)
		}
	}
	slices.Sort(ns)

	out := ns[:0]
	for i, n := range ns {
		if i == 0 || n != ns[i-1] {
			out = append(out, n)
		}
	}
	return out
}

// evalKnots reads one reed's fitted value at a note: linear between the knots and
// along the outermost segment past the ends. Linear here but flat in Curve.At,
// because it is only asked inside the recorded range, where the trend is evidence.
func evalKnots(ks []knot, n tuning.Note) float64 {
	switch len(ks) {
	case 0:
		return 0 // a reed nothing was recorded for: no goal, not a guess
	case 1:
		return ks[0].val
	}
	i, _ := slices.BinarySearchFunc(ks, n, func(k knot, n tuning.Note) int { return cmp.Compare(k.note, n) })
	switch {
	case i == 0:
		i = 1
	case i == len(ks):
		i = len(ks) - 1
	}
	a, b := ks[i-1], ks[i]
	t := float64(n-a.note) / float64(b.note-a.note)
	return a.val + t*(b.val-a.val)
}
