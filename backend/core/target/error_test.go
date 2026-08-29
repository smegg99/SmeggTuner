package target

import (
	"fmt"
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// gone marks a slot that did not sound, as the engine reports it.
const gone = meter.Missing

// measure is a reading as the engine makes one: a slot per deviation at the note's own octave,
// gone for a slot that did not sound, and the adjacent beats from the slots' frequencies.
func measure(note tuning.Note, a4 float64, devs ...float64) Reading {
	r := Reading{Note: note}
	for _, d := range devs {
		if d == gone {
			r.Reeds = append(r.Reeds, dsp.ReedMeasure{Dev: gone})
			continue
		}
		r.Reeds = append(r.Reeds, dsp.ReedMeasure{Present: true, Dev: d, Freq: meter.FrequencyHz(note.Freq(a4), d)})
	}
	r.Beats = beats(r.Reeds)
	return r
}

func beats(reeds []dsp.ReedMeasure) []dsp.BeatMeasure {
	var out []dsp.BeatMeasure
	for i := 1; i < len(reeds); i++ {
		b := dsp.BeatMeasure{Pair: fmt.Sprintf("%d-%d", i, i+1)}
		if reeds[i-1].Present && reeds[i].Present {
			b.Present, b.Hz = true, meter.HarmonicBeat(reeds[i-1].Freq, reeds[i].Freq, 0, true)
		}
		out = append(out, b)
	}
	return out
}

func musette(t *testing.T) *Curve {
	t.Helper()
	c := NewCurve("musette", 3)
	c.RefReed = 1
	mustSet(t, c, 60, 0, -6)
	mustSet(t, c, 60, 2, 6)
	mustSet(t, c, 84, 0, -12)
	mustSet(t, c, 84, 2, 12)
	return c
}

// No goal: the error is the deviation.
func TestErrorsWithoutCurve(t *testing.T) {
	m := measure(tuning.NoteA4, 440, -6.2, 0.4, 6.0)
	got := Errors(m, nil, 440, 0, nil)
	if len(got) != 3 {
		t.Fatalf("got %d rows, want 3", len(got))
	}
	for i, e := range got {
		if e.Reed != i {
			t.Fatalf("row %d is slot %d", i, e.Reed)
		}
		if e.Goal != 0 {
			t.Fatalf("slot %d: goal %v with no curve", i, e.Goal)
		}
		almost(t, e.Error, e.Curr, 1e-12, "error is the deviation")
	}
	if !got[1].InTol {
		t.Fatal("0.4 units is inside the default one unit tolerance")
	}
	if got[0].InTol || got[2].InTol {
		t.Fatal("6 units is not")
	}

	// An empty curve behaves the same as no curve.
	empty := Errors(m, NewCurve("", 3), 440, 0, nil)
	for i := range got {
		if empty[i] != got[i] {
			t.Fatalf("slot %d: empty curve %+v, no curve %+v", i, empty[i], got[i])
		}
	}
}

func TestErrorsAgainstCurve(t *testing.T) {
	c := musette(t)
	// Halfway between the anchors: slot 1 wants -9, slot 3 wants +9.
	m := measure(72, 440, -6.2, 0.5, 11.0)
	got := Errors(m, c, 440, 0, nil)

	almost(t, got[0].Goal, -9, 1e-9, "slot 1 goal")
	almost(t, got[0].Curr, -6.2, 1e-9, "slot 1 curr")
	almost(t, got[0].Error, 2.8, 1e-9, "slot 1 is 2.8 units sharp of its goal")
	almost(t, got[1].Goal, 0, 1e-12, "the reference reed is at pitch")
	almost(t, got[1].Error, 0.5, 1e-9, "slot 2 error")
	almost(t, got[2].Goal, 9, 1e-9, "slot 3 goal")
	almost(t, got[2].Error, 2.0, 1e-9, "slot 3 error")
	if !got[1].InTol || got[0].InTol || got[2].InTol {
		t.Fatalf("tolerance: %+v", got)
	}
}

// The two conventions must agree on how far the reed has to move.
func TestBothConventionsAreTheSameNumbers(t *testing.T) {
	c := musette(t)
	m := measure(72, 440, -6.2, 0.5, 11.0)
	for _, e := range Errors(m, c, 440, 0, nil) {
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
	mustSet(t, c, 40, 4, -2)
	mustSet(t, c, 80, 4, 6)

	m := measure(60, 440, 0.2, -0.5, 0.9, 0.1, 2.9)
	got := Errors(m, c, 440, 0, nil)
	if len(got) != 5 {
		t.Fatalf("got %d rows, want 5", len(got))
	}
	almost(t, got[4].Goal, 2, 1e-9, "slot 5 goal halfway between the anchors")
	almost(t, got[4].Error, 0.9, 1e-9, "slot 5 error")
	if !got[4].InTol {
		t.Fatal("0.9 units is inside the default tolerance")
	}
	for i := 0; i < 4; i++ {
		if got[i].Goal != 0 {
			t.Fatalf("slot %d: goal %v, want 0", i, got[i].Goal)
		}
	}
}

// The Hz fields mirror the deviations at the slot's pitch; ErrorHz is CurrHz - GoalHz.
func TestReedErrorCarriesHertz(t *testing.T) {
	c := musette(t)
	const a4 = 440
	m := measure(69, a4, -6.0, 0.0, 6.0) // A4: scale pitch is exactly a4

	for _, e := range Errors(m, c, a4, 0, nil) {
		wantCurr := meter.FrequencyHz(a4, e.Curr) - a4
		wantGoal := meter.FrequencyHz(a4, e.Goal) - a4
		almost(t, e.CurrHz, wantCurr, 1e-9, "the reed's deviation in hertz")
		almost(t, e.GoalHz, wantGoal, 1e-9, "the goal's deviation in hertz")
		almost(t, e.ErrorHz, wantCurr-wantGoal, 1e-9, "the distance between them")
	}
}
