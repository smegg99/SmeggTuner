package target

import (
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// A slot that did not sound has no row, and the slots after it keep their positions.
func TestAMissingSlotHasNoRow(t *testing.T) {
	c := musette(t)
	got := Errors(measure(72, 440, -6.2, gone, gone), c, 440, 0, nil)
	if len(got) != 1 {
		t.Fatalf("got %d rows, want 1", len(got))
	}
	almost(t, got[0].Goal, -9, 1e-9, "slot 1 still gets its goal")

	mid := Errors(measure(72, 440, gone, 0.5, 11), c, 440, 0, nil)
	if len(mid) != 2 || mid[0].Reed != 1 || mid[1].Reed != 2 {
		t.Fatalf("rows %+v, want slots 2 and 3 in place", mid)
	}
	almost(t, mid[1].Goal, 9, 1e-9, "slot 3 is read against its own column, not slot 2's")

	if rows := Errors(measure(72, 440), c, 440, 0, nil); len(rows) != 0 {
		t.Fatalf("a reading with no slots gave %d rows", len(rows))
	}
	if rows := Errors(Reading{}, c, 440, 0, nil); len(rows) != 0 {
		t.Fatalf("a heartbeat gave %d rows", len(rows))
	}
}

func TestMoreReedsThanCurve(t *testing.T) {
	c := NewCurve("single", 1)
	mustSet(t, c, tuning.NoteA4, 0, 2)

	got := Errors(measure(tuning.NoteA4, 440, 2.5, 3.0, 3.5), c, 440, 0, nil)
	if len(got) != 3 {
		t.Fatalf("got %d rows, want 3", len(got))
	}
	almost(t, got[0].Goal, 2, 1e-9, "slot 1 goal")
	if !got[0].HasGoal {
		t.Fatal("slot 1 has a goal")
	}
	for i := 1; i < 3; i++ {
		if got[i].Goal != 0 || got[i].HasGoal {
			t.Fatalf("slot %d: the curve says nothing about it, so it has no goal: %+v", i, got[i])
		}
		almost(t, got[i].Error, got[i].Curr, 1e-12, "and the error is the deviation")
	}
}

func TestTolerance(t *testing.T) {
	// Typed rather than measured: where the window closes only matters for a value exactly on it.
	r := Reading{
		Note: tuning.NoteA4,
		Reeds: []dsp.ReedMeasure{
			{Present: true, Dev: 1.0}, {Present: true, Dev: -1.0}, {Present: true, Dev: 1.01},
		},
	}
	got := Errors(r, nil, 440, 0, nil)
	if !got[0].InTol || !got[1].InTol {
		t.Fatal("one unit is inside the default one unit tolerance")
	}
	if got[2].InTol {
		t.Fatal("1.01 units is not")
	}
	if got := Errors(r, nil, 440, 3, nil); !got[2].InTol {
		t.Fatal("1.01 units is inside a 3 unit tolerance")
	}
	if got := Errors(r, nil, 440, 0.5, nil); got[0].InTol {
		t.Fatal("one unit is outside a half unit tolerance")
	}
}

// Curr is the engine's own deviation; only the Hz fields read the frequency, at the slot's octave.
func TestASlotIsMeasuredAtItsOwnOctave(t *testing.T) {
	const a4 = 440
	low := tuning.NoteA4.Freq(a4) / 2
	r := Reading{Note: tuning.NoteA4, Reeds: []dsp.ReedMeasure{
		{Present: true, Dev: 4, Octave: -12, Freq: meter.FrequencyHz(low, 4)},
	}}

	got := Errors(r, nil, a4, 0, nil)
	if len(got) != 1 {
		t.Fatalf("got %d rows", len(got))
	}
	almost(t, got[0].Curr, 4, 1e-12, "the deviation as measured")
	almost(t, got[0].CurrHz, low*0.0006*4, 1e-9, "4 units of a 220 Hz slot, not of 440 Hz")
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

// Past the curve's anchors there is no goal: blank, not zero, and the error is the deviation.
func TestNoGoalPastTheCurve(t *testing.T) {
	c := musette(t)
	m := measure(90, 440, -6, 0.5, 6)
	for _, e := range Errors(m, c, 440, 0, nil) {
		if e.HasGoal || e.Goal != 0 || e.Error != e.Curr || e.GoalHz != 0 {
			t.Fatalf("slot %d above the last anchor = %+v, want no goal", e.Reed, e)
		}
	}
	for _, b := range BeatErrors(m, c, 440, 0, nil) {
		if b.HasGoal || b.Goal != 0 || b.GoalHz != 0 || b.Error != b.Curr {
			t.Fatalf("beat %s above the last anchor = %+v, want no goal", b.Pair, b)
		}
	}
	for _, e := range Errors(measure(72, 440, -9, 0, 9), c, 440, 0, nil) {
		if !e.HasGoal {
			t.Fatalf("slot %d between the anchors has no goal", e.Reed)
		}
	}
}

// A transposed reading sounds above its key: its Hz are read at the sounding pitch, its key stays the note.
func TestATransposedReadingIsReadAtItsSoundingPitch(t *testing.T) {
	const a4 = 440
	sounding := (tuning.NoteA4 + 2).Freq(a4)
	r := Reading{Note: tuning.NoteA4, Transpose: 2, Reeds: []dsp.ReedMeasure{
		{Present: true, Dev: 5, Freq: meter.FrequencyHz(sounding, 5)},
	}}
	e := Errors(r, nil, a4, 0, nil)[0]
	almost(t, e.CurrHz, meter.FrequencyHz(sounding, 5)-sounding, 1e-9, "Hz at the sounding pitch")
	almost(t, e.Curr, 5, 1e-12, "the deviation is the engine's")
}
