package meter

import "testing"

func TestExtractPeaksSeedsTiesAndFloor(t *testing.T) {
	s := make([]float32, 20)
	s[5], s[9], s[13] = 8, 8, 1
	if got := ExtractPeaks(s, 3, 16); got != [3]int{5, 9, 0} {
		t.Fatalf("ties keep scan order and the floor drops weak ranks: %v", got)
	}
	flat := make([]float32, 20)
	if got := ExtractPeaks(flat, 4, 10); got != [3]int{4, 4, 4} {
		t.Fatalf("seeds survive on a flat band: %v", got)
	}
}

func TestAssignPeaks(t *testing.T) {
	targets := [3]float64{0, 10, -10}
	if got := AssignPeaks([3]int{205, 200, 0}, []int{1, 2}, targets, 200); got != [3]int{200, 205, 0} {
		t.Fatalf("full set sorts by bin: %v", got)
	}
	// One peak near the slot-2 target displaces the slot-1 default.
	if got := AssignPeaks([3]int{201, 0, 0}, []int{1, 2, 3}, targets, 200); got != [3]int{0, 201, 0} {
		t.Fatalf("single peak: %v", got)
	}
	if got := AssignPeaks([3]int{200, 0, 0}, []int{1, 2, 3}, targets, 200); got != [3]int{200, 0, 0} {
		t.Fatalf("single peak at slot-1 target: %v", got)
	}
}

func TestHistorySpanIsPhysical(t *testing.T) {
	if got := historySpan([8]float64{1, 2, 3, 4, 5, 6, 7, 8}); got != 7 {
		t.Fatalf("span %v", got)
	}
	if got := historySpan([8]float64{5, 0, 3, 4, 0, 120, 7, 8}); got != 120 {
		t.Fatalf("span after last zero %v", got)
	}
}

func TestHistoryEvaluate(t *testing.T) {
	h := History{LastNotes: [3]int{57, 0, 0}}
	for range 7 {
		if h.Evaluate([3]int{57, 0, 0}, [3]float64{1, 99, 99}, [3]int{5, -1, -1}) {
			t.Fatal("zeros in the fill keep the span at 100")
		}
	}
	if !h.Evaluate([3]int{57, 0, 0}, [3]float64{1, 99, 99}, [3]int{5, -1, -1}) || !h.Locked || !h.Acquired {
		t.Fatalf("full history must accept: %+v", h)
	}
	if h.Values[1] != [8]float64{} {
		t.Fatal("zero identities record zero")
	}
	h.Inhibited = true
	if h.Evaluate([3]int{60, 0, 0}, [3]float64{}, [3]int{4, -1, -1}) || h.Countdown != 4 || h.LastNotes != [3]int{57, 0, 0} {
		t.Fatalf("tuple change resets before the inhibit return: %+v", h)
	}
	h.Countdown = 1
	h.BeginAnalysis()
	if h.Locked || !h.Acquired {
		t.Fatal("countdown clears only the internal lock")
	}
}

// Steady readings of a note picked by fewer native events than Settle never lock; the next pick does.
func TestHistorySettleWaitsForNativePicks(t *testing.T) {
	h := History{Settle: 3}
	steady := func() bool { return h.Evaluate([3]int{57, 0, 0}, [3]float64{1, 1, 1}, [3]int{5, -1, -1}) }
	for pick := 1; pick < 3; pick++ {
		h.BeginAnalysis()
		for range 4 {
			if steady() {
				t.Fatalf("locked on native pick %d of 3", pick)
			}
		}
	}
	h.BeginAnalysis()
	if !steady() || !h.Locked {
		t.Fatalf("third pick with a full history must lock: %+v", h)
	}
}

func TestReadoutsOrdinary(t *testing.T) {
	r := NewReadouts(3)
	for i := range r.Panels {
		r.Panels[i] = Panel{Octave: 3, Enabled: true, Annotation: "x"}
	}
	r.Frequencies[1] = FrequencySlot{Deviation: 5, Frequency: 441, Position: 2, Valid: true}
	r.Ordinary(440, []float64{1, Missing, 60}, []float64{Missing, Missing, Missing}, false, true, true)
	if r.Frequencies[1].Valid || r.Frequencies[1].Frequency != 441 {
		t.Fatal("missing slot clears validity only")
	}
	if !r.Frequencies[0].Valid || r.Frequencies[2].Valid {
		t.Fatal("cache gate")
	}
	if r.Panels[1].Current != Missing || r.Panels[1].Decimals != 1 || r.Panels[0].Octave != 3 || r.Panels[0].Annotation != "" {
		t.Fatalf("panels %+v", r.Panels)
	}
	r.Reset()
	if r.Panels[0].Current != Missing || r.Panels[0].Octave != 0 || !r.Panels[0].Enabled || r.Frequencies[0].Valid || r.Frequencies[0].Frequency == 0 {
		t.Fatalf("reset %+v", r)
	}
}
