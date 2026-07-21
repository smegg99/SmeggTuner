package target

import (
	"testing"

	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/tuning"
)

func measure(note tuning.Note, a4 float64, devs ...float64) dsp.Measurement {
	m := dsp.Measurement{Note: note, ScalePitch: note.Freq(a4), ReedsSeparated: true}
	for _, d := range devs {
		m.Reeds = append(m.Reeds, dsp.ReedMeasure{
			Freq:     tuning.FreqAtCents(m.ScalePitch, d),
			DevCents: d,
		})
	}
	return m
}

func musette(t *testing.T) *Curve {
	t.Helper()
	c := NewCurve("musette", 3)
	c.RefReed = 1
	mustSet(t, c, 60, 0, -6, 440)
	mustSet(t, c, 60, 2, 6, 440)
	mustSet(t, c, 84, 0, -12, 440)
	mustSet(t, c, 84, 2, 12, 440)
	return c
}

// No goal: the error is the deviation.
func TestErrorsWithoutCurve(t *testing.T) {
	m := measure(tuning.NoteA4, 440, -6.2, 0.4, 6.0)
	got := Errors(m, nil, 440, 0)
	if len(got) != 3 {
		t.Fatalf("got %d rows, want 3", len(got))
	}
	for i, e := range got {
		if e.Reed != i {
			t.Fatalf("row %d is reed %d", i, e.Reed)
		}
		if e.Goal != 0 {
			t.Fatalf("reed %d: goal %v with no curve", i, e.Goal)
		}
		almost(t, e.Error, e.Curr, 1e-12, "error is the deviation")
	}
	if !got[1].InTol {
		t.Fatal("0.4 cents is inside the default 1 cent tolerance")
	}
	if got[0].InTol || got[2].InTol {
		t.Fatal("6 cents is not")
	}

	// An empty curve behaves the same as no curve.
	empty := Errors(m, NewCurve("", 3), 440, 0)
	for i := range got {
		if empty[i] != got[i] {
			t.Fatalf("reed %d: empty curve %+v, no curve %+v", i, empty[i], got[i])
		}
	}
}

func TestErrorsAgainstCurve(t *testing.T) {
	c := musette(t)
	// Halfway between the anchors: reed 1 wants -9, reed 3 wants +9.
	m := measure(72, 440, -6.2, 0.5, 11.0)
	got := Errors(m, c, 440, 0)

	almost(t, got[0].Goal, -9, 1e-9, "reed 1 goal")
	almost(t, got[0].Curr, -6.2, 1e-9, "reed 1 curr")
	almost(t, got[0].Error, 2.8, 1e-9, "reed 1 is 2.8 cents sharp of its goal")
	almost(t, got[1].Goal, 0, 1e-12, "the reference reed is at pitch")
	almost(t, got[1].Error, 0.5, 1e-9, "reed 2 error")
	almost(t, got[2].Goal, 9, 1e-9, "reed 3 goal")
	almost(t, got[2].Error, 2.0, 1e-9, "reed 3 error")
	if !got[1].InTol || got[0].InTol || got[2].InTol {
		t.Fatalf("tolerance: %+v", got)
	}
}

// The two conventions must agree on how far the reed has to move.
func TestBothConventionsAreTheSameNumbers(t *testing.T) {
	c := musette(t)
	m := measure(72, 440, -6.2, 0.5, 11.0)
	for _, e := range Errors(m, c, 440, 0) {
		scaleShow, scaleDrive := e.Display(RefScale)
		goalShow, goalDrive := e.Display(RefGoal)

		almost(t, scaleShow, e.Curr, 1e-12, "scale shows the deviation from the scale")
		almost(t, scaleDrive, e.Goal, 1e-12, "and you drive it to the curve")
		almost(t, goalShow, e.Error, 1e-12, "goal shows the distance from the curve")
		if goalDrive != 0 {
			t.Fatalf("goal convention drives to %v, not zero", goalDrive)
		}
		almost(t, scaleShow-scaleDrive, goalShow-goalDrive, 1e-12,
			"both conventions must ask for the same correction")
	}
}

func TestErrorsFiveReeds(t *testing.T) {
	c := NewCurve("bass", 5)
	mustSet(t, c, 40, 4, -2, 440)
	mustSet(t, c, 80, 4, 6, 440)

	m := measure(60, 440, 0.2, -0.5, 0.9, 0.1, 2.9)
	got := Errors(m, c, 440, 0)
	if len(got) != 5 {
		t.Fatalf("got %d rows, want 5", len(got))
	}
	almost(t, got[4].Goal, 2, 1e-9, "reed 5 goal halfway between the anchors")
	almost(t, got[4].Error, 0.9, 1e-9, "reed 5 error")
	if !got[4].InTol {
		t.Fatal("0.9 cents is inside the default tolerance")
	}
	for i := 0; i < 4; i++ {
		if got[i].Goal != 0 {
			t.Fatalf("reed %d: goal %v, want 0", i, got[i].Goal)
		}
	}
}

// The Hz fields mirror the cents deviations; ErrorHz is CurrHz - GoalHz.
func TestReedErrorCarriesHertz(t *testing.T) {
	c := musette(t)
	const a4 = 440
	m := measure(69, a4, -6.0, 0.0, 6.0) // A4: scale pitch is exactly a4

	for _, e := range Errors(m, c, a4, 0) {
		wantCurr := tuning.FreqAtCents(a4, e.Curr) - a4
		wantGoal := tuning.FreqAtCents(a4, e.Goal) - a4
		almost(t, e.CurrHz, wantCurr, 1e-9, "the reed's deviation in hertz")
		almost(t, e.GoalHz, wantGoal, 1e-9, "the goal's deviation in hertz")
		almost(t, e.ErrorHz, wantCurr-wantGoal, 1e-9, "the distance between them")
	}
}
