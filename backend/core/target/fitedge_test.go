package target

import (
	"errors"
	"testing"

	"smegg.me/smeggtuner/backend/core/tuning"
)

func TestFitEmptyPass(t *testing.T) {
	for _, rs := range [][]Reading{nil, {}} {
		if _, err := Fit(rs, 3, FitOptions{}); !errors.Is(err, ErrNoReadings) {
			t.Fatalf("empty pass: %v", err)
		}
	}
}

func TestFitReedCountRange(t *testing.T) {
	for _, n := range []int{0, -1, MaxReeds + 1} {
		if _, err := Fit(instrument(), n, FitOptions{}); err == nil {
			t.Fatalf("reed count %d was accepted", n)
		}
	}
	for _, n := range []int{MinReeds, MaxReeds} {
		if _, err := Fit(instrument(), n, FitOptions{}); err != nil {
			t.Fatalf("reed count %d: %v", n, err)
		}
	}
}

// One note is not a trend, but it is still a curve: one anchor at what was measured, and no goal elsewhere.
func TestFitSingleNote(t *testing.T) {
	res := mustFit(t, []Reading{reading(69, -8, 0, 8)}, 3)
	if len(res.Curve.Anchors) != 1 {
		t.Fatalf("%d anchors from one note", len(res.Curve.Anchors))
	}
	got := res.Curve.At(69)
	almost(t, got[0], -8, 1e-9, "the note that was recorded")
	almost(t, got[2], 8, 1e-9, "the note that was recorded")
	for _, n := range []tuning.Note{tuning.MinNote, 68, 70, tuning.MaxNote} {
		noGoal(t, res.Curve, n, "nothing was recorded there")
	}
}

// Fit reads the slots exactly as Errors does.
func TestFitAgreesWithErrorsOnWhatWasRecorded(t *testing.T) {
	res := mustFit(t, instrument(), 3)
	for n := fitLo; n <= fitHi; n++ {
		m := measure(n, fitA4, -10, 0, ramp(n))
		for _, e := range Errors(m, res.Curve, fitA4, 0, nil) {
			if !e.InTol {
				t.Fatalf("note %d slot %d: the instrument reads %.2f units out of its "+
					"own fitted curve", n, e.Reed+1, e.Error)
			}
		}
	}
}
