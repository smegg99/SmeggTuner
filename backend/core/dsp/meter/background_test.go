package meter

import (
	"math"
	"slices"
	"testing"
)

// bgNote places band at bin 1 between two neighbours; a zero Q becomes the constructor's one.
func bgNote(left float32, band []float32, right float32, s NoteState) (NoteState, []float32) {
	s.Lower = 1
	if s.Reference == nil {
		s.Reference = make([]float32, len(band))
	}
	if s.Q == 0 {
		s.Q = 1
	}
	return s, append(append([]float32{left}, band...), right)
}

func bgUpdate(band []float32, s NoteState, selected bool) (NoteState, NoteStatistics, bool) {
	s, spectrum := bgNote(0, band, 0, s)
	stats, raw := s.Update(spectrum, selected)
	return s, stats, raw
}

func TestAdaptReferenceIsAsymmetricBinary32(t *testing.T) {
	ref, tie, still := []float32{0, 200, 17, 0}, []float32{1}, []float32{1}
	AdaptReference(ref, []float32{150, 0, 17, 9}) // 9/150 is not 9*F(1/150)
	// 99/150 rounds first; adding one then ties to the even upper neighbour.
	AdaptReference(tie, []float32{100})
	AdaptReference(still, []float32{math.Nextafter32(1, 2)})
	AdaptReference(nil, nil)
	if !slices.Equal(ref, []float32{1, 199, 17, 0.05999999865889549}) || math.Float32bits(tie[0]) != 0x3fd47ae2 || still[0] != 1 {
		t.Fatalf("rise /150, fall /200, rounded step, stagnation: %v %08x %v", ref, math.Float32bits(tie[0]), still[0])
	}
}

func TestAggregateBackground(t *testing.T) {
	runs := func(pairs ...float32) (m [ActiveNotes]float32) {
		i := 0
		for p := 0; p < len(pairs); p += 2 {
			for range int(pairs[p]) {
				m[i], i = pairs[p+1], i+1
			}
		}
		return m
	}
	for i, c := range []struct {
		in   [ActiveNotes]float32
		want float32
	}{
		{runs(105, 0), 0}, {runs(105, 4), 4},
		{runs(26, 0, 79, 105), 79}, {runs(27, 0, 78, 105), 0}, {runs(27, 1, 78, 105), 1},
		{runs(27, 1, 77, 5, 1, 8), 4}, // all 27 sit exactly on the strict cutoff
		{runs(79, 1<<20, 26, 1), 788928.625}, {runs(26, 1, 79, 1<<20), 788928.8125},
	} {
		if got := AggregateBackground(c.in); got != c.want {
			t.Errorf("case %d: got %v want %v", i, got, c.want)
		}
	}
}

func TestUpdateScansInclusiveBandAndNeighbours(t *testing.T) {
	s := NewNoteState(10, 15)
	stats, _ := s.Update([]float32{9: 0, 3, 3, 0, 1, 0, 3, 4}, false)
	if want := (NoteStatistics{3, 10, 0, 12, 1.6666666269302368, 2, 3, 2, 1.5}); stats != want || !s.Foreground {
		t.Fatalf("first ties, left-edge plateau, ratio 1.5 admitted: %+v %v", stats, s.Foreground)
	}
	for _, c := range []struct{ left, right, want float32 }{{0, 0, 2}, {3, 0, 1}, {0, 2, 1}, {3, 2, 0}} {
		n, spectrum := bgNote(c.left, []float32{3, 0, 1}, c.right, NoteState{})
		if stats, _ := n.Update(spectrum, false); float32(stats.PeakCount) != c.want {
			t.Errorf("neighbours %v: peaks %d", c, stats.PeakCount)
		}
	}
	n, spectrum := bgNote(2, []float32{2, 2, 2}, 2, NoteState{EntrySnapshot: 7, EntryKnown: true})
	if stats, raw := n.Update(spectrum, false); stats.PeakCount != 0 || stats.PeakRatio != 0 || raw || n.EntrySnapshot != 7 || n.Threshold <= 0 {
		t.Fatalf("a plateau has no peak: %+v raw %v state %+v", stats, raw, n)
	}
}

func TestEntrySnapshotPrecedesRatioVeto(t *testing.T) {
	s, _, raw := bgUpdate([]float32{100}, NoteState{}, false)
	if !raw || s.Foreground || !s.EntryKnown || s.EntrySnapshot != 2 {
		t.Fatalf("a sole peak is vetoed after the snapshot write: %+v", s)
	}
	again, _, _ := bgUpdate([]float32{100}, s, false)
	held, _, _ := bgUpdate([]float32{100}, NoteState{Foreground: true, EntrySnapshot: 17, EntryKnown: true}, false)
	quiet, _, _ := bgUpdate([]float32{0}, NoteState{}, false)
	if again.EntrySnapshot != 4 || held.EntrySnapshot != 17 || held.Foreground || quiet.EntryKnown {
		t.Fatalf("vetoed entries rewrite, old foreground blocks, no entry stays unknown: %+v %+v %+v", again, held, quiet)
	}
}

func TestThresholdBoundaries(t *testing.T) {
	_, at, _ := bgUpdate([]float32{3, 0, 1}, NoteState{}, false)
	below, _, _ := bgUpdate([]float32{3, 0, 1 + 1.0/(1<<21)}, NoteState{}, false)
	edge, _, raw := bgUpdate([]float32{100, 0, 20}, NoteState{Threshold: 98}, false)
	over, _, _ := bgUpdate([]float32{100, 0, 20}, NoteState{Threshold: 97}, false)
	if at.PeakRatio != 1.5 || below.Foreground || edge.Threshold != 100 || raw || !over.Foreground {
		t.Fatalf("ratio gate admits 1.5, amplitude gate is strict on the new threshold: %v %v %v", at.PeakRatio, edge.Threshold, over.Foreground)
	}
	rise, _, _ := bgUpdate([]float32{100}, NoteState{}, false)
	half, _, _ := bgUpdate([]float32{100}, NoteState{}, true)
	decay, _, _ := bgUpdate([]float32{3}, NoteState{Threshold: 30}, true)
	if rise.Threshold != 2 || half.Threshold != 1 || decay.Threshold != 20 {
		t.Fatalf("rise /50, selected /100, decay on equality: %v %v %v", rise.Threshold, half.Threshold, decay.Threshold)
	}
	// F(0.1)*10 rounds to 1 in binary32, so the comparison decays; S - S/3 is not F(2/3).
	tenth := float32(0.1)
	if d, _, _ := bgUpdate([]float32{tenth}, NoteState{Reference: []float32{tenth}, Threshold: 1}, false); d.Threshold != 0.6666666269302368 {
		t.Fatalf("decay order: %v", d.Threshold)
	}
}

func TestQIsConvertedBeforeMultiplying(t *testing.T) {
	if two, _, _ := bgUpdate([]float32{3, 0, 1}, NoteState{Threshold: 4, Q: 2}, false); !two.Foreground || two.Q != 2 {
		t.Fatalf("q=2 doubles the peak: %+v", two)
	}
	for q, want := range map[int32]bool{1<<24 + 1: false, 1<<24 + 2: true} {
		if _, _, raw := bgUpdate([]float32{1}, NoteState{Reference: []float32{1}, Threshold: 25165824, Q: q}, false); raw != want {
			t.Errorf("q %d: raw %v", q, raw)
		}
	}
	s, spectrum := bgNote(0, []float32{150, 0, 50}, 0, NoteState{Foreground: true, EntrySnapshot: 7, EntryKnown: true})
	s.Q = 0
	if _, raw := s.Update(spectrum, false); raw || s.Foreground || s.Threshold != 3 || s.Reference[0] != 1 || s.EntrySnapshot != 7 {
		t.Fatalf("q=0 clears foreground on the next update while adaptation continues: %+v", s)
	}
}

func TestStatisticsIgnoreReferenceAndKeepOrder(t *testing.T) {
	loud, a, _ := bgUpdate([]float32{100, 0, 20}, NoteState{Reference: []float32{200, 200, 200}}, false)
	quiet, b, _ := bgUpdate([]float32{100, 0, 20}, NoteState{}, false)
	if a != b || loud.Threshold != quiet.Threshold || loud.Reference[0] != 199.5 || loud.PreviousMaximum != 100 {
		t.Fatalf("the reference is not subtracted: %+v %+v", a, b)
	}
	s, spectrum := bgNote(1<<20, []float32{1 << 20, 0.0625, 0.0625}, 0.0625, NoteState{})
	if stats, _ := s.Update(spectrum, false); stats.Mean != 349525.34375 {
		t.Fatalf("sequential binary32 sum: %v", stats.Mean)
	}
	s.Q, s.Threshold = 0, 9
	if s.Reconfigure(4, 8); len(s.Reference) != 5 || s.Lower != 4 || s.Q != 0 || s.Threshold != 9 || s.PreviousMaximum != 1<<20 {
		t.Fatalf("reconfiguration only replaces the band: %+v", s)
	}
}
