package target

import (
	"errors"
	"testing"

	"smegg.me/smeggtuner/core/tuning"
)

func TestFitEmptyPass(t *testing.T) {
	for _, rs := range [][]Reading{nil, {}} {
		if _, err := Fit(rs, 3, fitA4, FitOptions{}); !errors.Is(err, ErrNoReadings) {
			t.Fatalf("empty pass: %v", err)
		}
	}
}

func TestFitReedCountRange(t *testing.T) {
	for _, n := range []int{0, -1, MaxReeds + 1} {
		if _, err := Fit(instrument(), n, fitA4, FitOptions{}); err == nil {
			t.Fatalf("reed count %d was accepted", n)
		}
	}
	for _, n := range []int{MinReeds, MaxReeds} {
		if _, err := Fit(instrument(), n, fitA4, FitOptions{}); err != nil {
			t.Fatalf("reed count %d: %v", n, err)
		}
	}
}

// One note is not a trend, but it is still a curve: flat, at what was measured.
func TestFitSingleNote(t *testing.T) {
	res := mustFit(t, []Reading{reading(69, -8, 0, 8)}, 3)
	if len(res.Curve.Anchors) != 1 {
		t.Fatalf("%d anchors from one note", len(res.Curve.Anchors))
	}
	for _, n := range []tuning.Note{tuning.MinNote, 69, tuning.MaxNote} {
		got := res.Curve.At(n)
		almost(t, got[0], -8, 1e-9, "held flat across the keyboard")
		almost(t, got[2], 8, 1e-9, "held flat across the keyboard")
	}
}

// Fit recovers values from the frequencies exactly as Errors does.
func TestFitAgreesWithErrorsOnWhatWasRecorded(t *testing.T) {
	res := mustFit(t, instrument(), 3)
	for n := fitLo; n <= fitHi; n++ {
		m := measure(n, fitA4, -10, 0, ramp(n))
		for _, e := range Errors(m, res.Curve, fitA4, 0) {
			if !e.InTol {
				t.Fatalf("note %d reed %d: the instrument reads %.2f cents out of its "+
					"own fitted curve", n, e.Reed+1, e.Error)
			}
		}
	}
}
