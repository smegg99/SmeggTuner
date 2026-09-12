package meter

import "testing"

// coarseFixture has five-bin bands 2+7i..6+7i and two peaks, 3 and 1, in note 10's band.
func coarseFixture() (CoarseState, []float32) {
	var bands [ActiveNotes]NoteBand
	for i := range bands {
		bands[i] = NoteBand{LowerBin: 2 + 7*i, UpperBin: 6 + 7*i}
	}
	s := NewCoarseState(&bands)
	spectrum := make([]float32, 740)
	copy(spectrum[s.Notes[6].Lower:], []float32{0, 3, 0, 1, 0})
	return s, spectrum
}

func TestCoarseNewStateIsIdleAndFresh(t *testing.T) {
	s, _ := coarseFixture()
	if s.Learning != NewLearning() || s.Background != 0 || s.Scores != [110]float32{} {
		t.Fatalf("construction: %+v %v %v", s.Learning, s.Background, s.Scores)
	}
	for i, n := range s.Notes {
		if n.Lower != 2+7*i || len(n.Reference) != 5 || n.Q != 1 || n.Threshold != 0 || n.Foreground || n.EntryKnown {
			t.Fatalf("note %d: %+v", i+FirstActiveNote, n)
		}
	}
}

func TestCoarseSelectionPrecedesRefresh(t *testing.T) {
	s, spectrum := coarseFixture()
	if f := s.Refresh(spectrum, fixtureUnset()); f.Selected != 0 || f.LearningStep != 0 {
		t.Fatalf("the first refresh selects from the empty stored table: %+v", f)
	}
	if s.Scores[10] != 3 || !s.Notes[6].Foreground {
		t.Fatalf("refresh stores the score and foreground: %v %+v", s.Scores[10], s.Notes[6])
	}
	if f := s.Refresh(make([]float32, 740), fixtureUnset()); f.Selected != 10 {
		t.Fatalf("silence still selects from the stored table: %+v", f)
	}
	if s.Notes[6].Foreground {
		t.Fatal("the silent refresh clears foreground afterwards")
	}
	if f := s.Refresh(make([]float32, 740), fixtureUnset()); f.Selected != 0 {
		t.Fatalf("the next frame sees the cleared table: %+v", f)
	}
}

func TestCoarseLearningStepEightSelectsBeforeSeededRefresh(t *testing.T) {
	s, spectrum := coarseFixture()
	s.Refresh(spectrum, fixtureUnset())
	s.Learning = Learning{Request: true, Counter: 7}
	f := s.Refresh(spectrum, fixtureUnset())
	if f.LearningStep != 8 || f.Selected != 10 || s.Learning.Request {
		t.Fatalf("step eight ends learning and selects: %+v %+v", f, s.Learning)
	}
	var score, divisor, seed float32 = 3, 100, 12
	if s.Notes[6].Foreground || s.Notes[6].Threshold != score/divisor+seed {
		t.Fatalf("seeded threshold, then the selected-note rise: %+v", s.Notes[6])
	}
}

func TestCoarseAcquisitionSkipsManualBypass(t *testing.T) {
	s, spectrum := coarseFixture()
	s.Learning = Learning{Request: true, Counter: -1}
	d := Nomination{Octave: 4, PitchClass: 9, MeasureContinuously: true}
	if f := s.Refresh(spectrum, d); f.Selected != 0 || f.LearningStep != 1 {
		t.Fatalf("acquisition skips selection entirely: %+v", f)
	}
	s.Learning = NewLearning()
	if f := s.Refresh(spectrum, d); f.Selected != 57 {
		t.Fatalf("idle learning lets the bypass through: %+v", f)
	}
}

func TestCoarseQZeroSuppressesForeground(t *testing.T) {
	s, spectrum := coarseFixture()
	s.Notes[6].Q = 0
	s.Refresh(spectrum, fixtureUnset())
	if s.Notes[6].Q != 0 || s.Notes[6].Foreground || s.Scores[10] != 3 {
		t.Fatalf("q zero survives and fails the threshold: %+v %v", s.Notes[6], s.Scores[10])
	}
}

func TestCoarseKeepsBoundaryNeighbours(t *testing.T) {
	s, spectrum := coarseFixture()
	s.Scores[3], s.Scores[109] = 9, 13
	s.Refresh(spectrum, fixtureUnset())
	if s.Scores[3] != 9 || s.Scores[109] != 13 {
		t.Fatalf("boundary scores are never refreshed: %v %v", s.Scores[3], s.Scores[109])
	}
}
