package meter

import (
	"math"
	"slices"
	"testing"
)

// bgPopulation repeats one band for every active note, separated by zero bins.
func bgPopulation(band []float32, s NoteState) (*[ActiveNotes]NoteState, []float32) {
	var notes [ActiveNotes]NoteState
	spectrum := []float32{0}
	for i := range notes {
		notes[i] = s
		notes[i].Lower, notes[i].Reference = len(spectrum), slices.Clone(s.Reference)
		spectrum = append(append(spectrum, band...), 0)
	}
	return &notes, spectrum
}

func TestLearningCounterAndRequest(t *testing.T) {
	notes, spectrum := bgPopulation([]float32{150}, NewNoteState(0, 0))
	idle, start := NewLearning(), Learning{Request: true, Counter: -1}
	if idle.Prepare(notes, spectrum, 0) != 0 || notes[0].Reference[0] != 0 || !idle.SelectionAllowed() {
		t.Fatal("idle learning must not touch notes")
	}
	if step := start.Prepare(notes, spectrum, 0); step != 1 || start != (Learning{Request: true, Counter: 1}) || start.SelectionAllowed() {
		t.Fatalf("a request starts at step 1 and keeps blocking selection: %d %+v", step, start)
	}
	for _, request := range []bool{false, true} {
		active, last := Learning{Request: request, Counter: 4}, Learning{Request: request, Counter: 7}
		active.Prepare(notes, spectrum, 0)
		if step := last.Prepare(notes, spectrum, 0); active != (Learning{Request: request, Counter: 5}) || step != 8 || last != (Learning{Counter: -1, Done: true}) {
			t.Fatalf("request %v: repeats never restart, clearing never aborts, step 8 ends and marks the pass done: %+v %+v", request, active, last)
		}
	}
	if zero := (Learning{Counter: 0}); zero.Prepare(notes, spectrum, 0) != 1 {
		t.Fatal("an active counter advances without a request")
	}
}

func TestLearningSeedsFromPreviousState(t *testing.T) {
	kept := NoteState{Reference: []float32{1, 2}, Threshold: 17, Foreground: true, EntrySnapshot: 9, EntryKnown: true, PreviousMaximum: 23}
	notes, spectrum := bgPopulation([]float32{10, 20}, kept)
	copying := Learning{Request: true, Counter: 2}
	if copying.Prepare(notes, spectrum, 4); !slices.Equal(notes[104].Reference, []float32{10, 20}) || notes[104].Threshold != 17 {
		t.Fatalf("step 3 copies the band and nothing else: %+v", notes[104])
	}
	notes, spectrum = bgPopulation([]float32{151}, NoteState{Reference: []float32{1}, Threshold: 999, PreviousMaximum: 2, Q: 1})
	seeding := Learning{Request: true, Counter: 3}
	if seeding.Prepare(notes, spectrum, 3); notes[0].Reference[0] != 1 || notes[0].Threshold != 53 {
		t.Fatalf("minimum and seed 4*A + 15*G from previous values: %+v", notes[0])
	}
	if notes[0].Update(spectrum, false); notes[0].Reference[0] != 2 || notes[0].Threshold != 56.02000045776367 {
		t.Fatalf("adaptation lifts the minimum on the same frame: %+v", notes[0])
	}
	x := float32(0.3)
	notes, spectrum = bgPopulation([]float32{0}, NoteState{Reference: []float32{0}})
	for i := range notes {
		notes[i].PreviousMaximum = float32(i)
	}
	notes[0].PreviousMaximum, seeding = x, Learning{Request: true, Counter: 3}
	if seeding.Prepare(notes, spectrum, x); notes[0].Threshold != 5.699999809265137 || notes[104].Threshold != 420.5 {
		t.Fatalf("products round before the sum, per note: %v %v", notes[0].Threshold, notes[104].Threshold)
	}
}

func TestLearningMinimumKeepsLowerBinsAndSignedZeroTies(t *testing.T) {
	negative := float32(math.Copysign(0, -1))
	notes, spectrum := bgPopulation([]float32{10, 20, 0, negative}, NewNoteState(0, 3))
	l := Learning{Request: true, Counter: 2}
	l.Prepare(notes, spectrum, 0)
	_, lower := bgPopulation([]float32{5, 30, negative, 0}, NoteState{})
	l.Prepare(notes, lower, 0)
	got := notes[104].Reference
	if !slices.Equal(got, []float32{5, 20, 0, 0}) || math.Signbit(float64(got[2])) || !math.Signbit(float64(got[3])) {
		t.Fatalf("step 4 keeps the lower bin, the old higher bin and the old signed zero: %v", got)
	}
}

func TestLearningComposesWithUpdates(t *testing.T) {
	notes, spectrum := bgPopulation([]float32{150, 0, 50}, NewNoteState(0, 2))
	l, background := Learning{Request: true, Counter: -1}, float32(0)
	for step := 1; step <= 8; step++ {
		if got := l.Prepare(notes, spectrum, background); got != step || l.SelectionAllowed() != (step == 8) {
			t.Fatalf("step %d: got %d", step, got)
		}
		var maxima [ActiveNotes]float32
		for i := range notes {
			stats, _ := notes[i].Update(spectrum, false)
			maxima[i] = stats.Maximum
		}
		background = AggregateBackground(maxima)
		n, bad := notes[0], background != 150
		switch {
		case step == 1:
			bad = bad || n.Reference[0] != 1
		case step == 3:
			bad = bad || !slices.Equal(n.Reference, []float32{150, 0, 50}) || n.Threshold != 9
		case step >= 4:
			bad = bad || n.Threshold != 1900
		}
		if bad {
			t.Fatalf("step %d: %+v background %v", step, n, background)
		}
	}
	if l.Request = true; l.Prepare(notes, spectrum, background) != 1 {
		t.Fatal("relearning restarts from step 1")
	}
	notes, spectrum = bgPopulation([]float32{100}, NewNoteState(0, 0))
	last := Learning{Request: false, Counter: 7}
	last.Prepare(notes, spectrum, 0)
	notes[0].Update(spectrum, last.SelectionAllowed())
	notes[1].Update(spectrum, false)
	if notes[0].Threshold != 1 || notes[1].Threshold != 2 {
		t.Fatalf("step 8 clears the request before selection, so the rise is /100: %v %v", notes[0].Threshold, notes[1].Threshold)
	}
}
