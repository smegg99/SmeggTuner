package target

import (
	"math"
	"testing"

	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/tuning"
)

// A reed that did not sound leaves the measurement shorter than the curve.
func TestFewerReedsThanCurve(t *testing.T) {
	c := musette(t)
	m := measure(72, 440, -6.2)
	got := Errors(m, c, 440, 0)
	if len(got) != 1 {
		t.Fatalf("got %d rows, want 1", len(got))
	}
	almost(t, got[0].Goal, -9, 1e-9, "reed 1 still gets its goal")

	if rows := Errors(measure(72, 440), c, 440, 0); len(rows) != 0 {
		t.Fatalf("a measurement with no reeds gave %d rows", len(rows))
	}
	if rows := Errors(dsp.Measurement{State: dsp.StateTooQuiet}, c, 440, 0); len(rows) != 0 {
		t.Fatalf("a heartbeat gave %d rows", len(rows))
	}
}

func TestMoreReedsThanCurve(t *testing.T) {
	c := NewCurve("single", 1)
	mustSet(t, c, tuning.NoteA4, 0, 2, 440)

	got := Errors(measure(tuning.NoteA4, 440, 2.5, 3.0, 3.5), c, 440, 0)
	if len(got) != 3 {
		t.Fatalf("got %d rows, want 3", len(got))
	}
	almost(t, got[0].Goal, 2, 1e-9, "reed 1 goal")
	for i := 1; i < 3; i++ {
		if got[i].Goal != 0 {
			t.Fatalf("reed %d: the curve says nothing about it, so goal must be 0", i)
		}
		almost(t, got[i].Error, got[i].Curr, 1e-12, "and the error is the deviation")
	}
}

func TestTolerance(t *testing.T) {
	// Typed rather than measured: where the window closes only matters for a value exactly on it.
	m := dsp.Measurement{
		Note: tuning.NoteA4,
		Reeds: []dsp.ReedMeasure{
			{DevCents: 1.0}, {DevCents: -1.0}, {DevCents: 1.01},
		},
	}
	got := Errors(m, nil, 440, 0)
	if !got[0].InTol || !got[1].InTol {
		t.Fatal("one cent is inside the default one cent tolerance")
	}
	if got[2].InTol {
		t.Fatal("1.01 cents is not")
	}
	if got := Errors(m, nil, 440, 3); !got[2].InTol {
		t.Fatal("1.01 cents is inside a 3 cent tolerance")
	}
	if got := Errors(m, nil, 440, 0.5); got[0].InTol {
		t.Fatal("one cent is outside a half cent tolerance")
	}
}

// A take read back from a saved session has frequencies but no scale pitch: it
// is measured against the session's own A4, so a pass survives the A4 slider.
func TestErrorsFromRebuiltTake(t *testing.T) {
	const a4 = 442
	live := measure(tuning.NoteA4, a4, 4.0)
	rebuilt := dsp.Measurement{Note: tuning.NoteA4, Reeds: []dsp.ReedMeasure{{Freq: live.Reeds[0].Freq}}}

	got := Errors(rebuilt, nil, a4, 0)
	if len(got) != 1 {
		t.Fatalf("got %d rows", len(got))
	}
	almost(t, got[0].Curr, 4.0, 1e-9, "curr recovered from the frequency")

	at440 := Errors(rebuilt, nil, 440, 0)
	almost(t, at440[0].Curr, 4.0+tuning.Cents(442, 440), 1e-9, "curr at the wrong A4")
}

// A hand-edited Curr has no frequency behind it and must stand as typed.
func TestManualCurrStandsAsTyped(t *testing.T) {
	m := dsp.Measurement{
		Note:  tuning.NoteA4,
		Reeds: []dsp.ReedMeasure{{DevCents: -3.5}},
	}
	got := Errors(m, nil, 440, 0)
	almost(t, got[0].Curr, -3.5, 1e-12, "typed value")
	almost(t, got[0].Error, -3.5, 1e-12, "and its error")
	if math.IsNaN(got[0].Curr) {
		t.Fatal("NaN")
	}
}

// A non-positive window falls back to the default, so a zero from config never means "nothing is ever in tune".
func TestTolerances(t *testing.T) {
	cases := []struct {
		reed, beat         float64
		wantReed, wantBeat float64
	}{
		{0, 0, DefaultTolerance, DefaultBeatTolerance},
		{-1, -1, DefaultTolerance, DefaultBeatTolerance},
		{0.5, 2, 0.5, 2},
		{0, 2, DefaultTolerance, 2},
		{0.5, 0, 0.5, DefaultBeatTolerance},
	}
	for _, c := range cases {
		reed, beat := Tolerances(c.reed, c.beat)
		if reed != c.wantReed || beat != c.wantBeat {
			t.Fatalf("Tolerances(%v, %v) = %v / %v, want %v / %v",
				c.reed, c.beat, reed, beat, c.wantReed, c.wantBeat)
		}
	}
}
