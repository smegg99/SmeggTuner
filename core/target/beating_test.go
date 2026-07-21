package target

import (
	"fmt"
	"testing"

	"smegg.me/smeggtuner/core/tuning"
)

// hzAtNote reads a stored (cents) goal back as the Hz offset it was authored as.
func hzAtNote(t *testing.T, c *Curve, note tuning.Note, a4 float64) []float64 {
	t.Helper()
	goal := c.At(note)
	out := make([]float64, len(goal))
	for i, cents := range goal {
		out[i] = HzFromCents(note, cents, a4)
	}
	return out
}

func mustBeat(t *testing.T, c *Curve, note tuning.Note, value, a4 float64) {
	t.Helper()
	if err := c.SetBeating(note, value, a4); err != nil {
		t.Fatalf("SetBeating(%d, %v): %v", note, value, err)
	}
}

// SetBeating derives the minus / zero / plus of a musette from one typed value.
func TestSetBeatingDerivesTheReeds(t *testing.T) {
	const a4 = 440
	c := NewCurve("musette", 3)
	c.Unit = UnitHz
	c.RefReed = 1

	mustBeat(t, c, tuning.NoteA4, 2, a4)
	got := hzAtNote(t, c, tuning.NoteA4, a4)
	want := []float64{-2, 0, 2}
	for r := range want {
		almost(t, got[r], want[r], 1e-9, "2 Hz beating, reed 2 on zero")
	}
	almost(t, got[1]-got[0], 2, 1e-9, "reeds 1 and 2 beat at what was typed")
	almost(t, got[2]-got[1], 2, 1e-9, "reeds 2 and 3 beat at what was typed")
	almost(t, got[2]-got[0], 4, 1e-9, "the outermost pair is two steps wide")
	almost(t, c.Beating(tuning.NoteA4, a4), 2, 1e-9, "Beating reads it straight back")

	k := NewCurve("cents", 3)
	k.RefReed = 1
	mustBeat(t, k, 60, 9, a4)
	for r, want := range []float64{-9, 0, 9} {
		almost(t, k.At(60)[r], want, 1e-12, "9 cents of beating")
	}
	almost(t, k.Beating(60, a4), 9, 1e-12, "Beating in cents")
}

// RefReed decides which reed the tremolo hangs off; never assume the middle one.
func TestSetBeatingHonoursRefReed(t *testing.T) {
	const a4 = 442
	cases := []struct {
		ref  int
		want []float64
	}{
		{0, []float64{0, 2, 4}},
		{1, []float64{-2, 0, 2}},
		{2, []float64{-4, -2, 0}},
		{NoRefReed, []float64{0, 2, 4}}, // no reed at pitch: reed 1 anchors it
	}
	for _, tc := range cases {
		c := NewCurve("musette", 3)
		c.Unit = UnitHz
		c.RefReed = tc.ref
		mustBeat(t, c, tuning.NoteA4, 2, a4)

		got := hzAtNote(t, c, tuning.NoteA4, a4)
		for r := range tc.want {
			almost(t, got[r], tc.want[r], 1e-9, fmt.Sprintf("ref reed %d", tc.ref))
		}
		almost(t, got[2]-got[0], 4, 1e-9, "the width is the same wherever it hangs")
	}
}

// Asymmetry moves the reference reed inside the tremolo without changing its width.
func TestAsymmetrySplitsTheBeating(t *testing.T) {
	const a4 = 440
	cases := []struct {
		asym float64
		want []float64
	}{
		{0, []float64{-2, 0, 2}},
		{100, []float64{0, 0, 4}},
		{-100, []float64{-4, 0, 0}},
		{50, []float64{-1, 0, 3}},
		{-50, []float64{-3, 0, 1}},
	}
	for _, tc := range cases {
		c := NewCurve("musette", 3)
		c.Unit = UnitHz
		c.RefReed = 1
		c.Asymmetry = tc.asym
		mustBeat(t, c, tuning.NoteA4, 2, a4)

		got := hzAtNote(t, c, tuning.NoteA4, a4)
		for r := range tc.want {
			almost(t, got[r], tc.want[r], 1e-9, "asymmetry")
		}
		almost(t, got[2]-got[0], 4, 1e-9, "asymmetry may not change the width")
		almost(t, got[1], 0, 1e-12, "the reference reed stays at pitch")
		if err := c.Validate(); err != nil {
			t.Fatalf("asymmetry %v: %v", tc.asym, err)
		}
	}

	// Reference reed at the end: no reeds to divide, so the percentage does nothing.
	for _, asym := range []float64{-100, 0, 100} {
		c := NewCurve("musette", 3)
		c.Unit = UnitHz
		c.RefReed = 0
		c.Asymmetry = asym
		mustBeat(t, c, tuning.NoteA4, 2, a4)
		for r, want := range []float64{0, 2, 4} {
			almost(t, hzAtNote(t, c, tuning.NoteA4, a4)[r], want, 1e-9,
				"reed 1 on zero: nothing below it to divide")
		}
	}

	bad := NewCurve("musette", 3)
	bad.Asymmetry = 101
	if err := bad.Validate(); err == nil {
		t.Fatal("asymmetry past 100 percent must not validate")
	}
}

// Any reed count spreads evenly, one step of the typed value between neighbours.
func TestSetBeatingAnyReedCount(t *testing.T) {
	const a4 = 440

	five := NewCurve("bass", 5)
	five.Unit = UnitHz
	five.RefReed = 0
	mustBeat(t, five, 60, 2, a4)
	for r, want := range []float64{0, 2, 4, 6, 8} {
		almost(t, hzAtNote(t, five, 60, a4)[r], want, 1e-9, "five reeds, reed 1 on zero")
	}

	five.RefReed = 2
	mustBeat(t, five, 60, 2, a4)
	for r, want := range []float64{-4, -2, 0, 2, 4} {
		almost(t, hzAtNote(t, five, 60, a4)[r], want, 1e-9, "five reeds, reed 3 on zero")
	}
	almost(t, five.Beating(60, a4), 2, 1e-9, "five reeds beat at what was typed")

	two := NewCurve("pair", 2)
	two.Unit = UnitHz
	two.RefReed = 0
	mustBeat(t, two, 60, 3, a4)
	for r, want := range []float64{0, 3} {
		almost(t, hzAtNote(t, two, 60, a4)[r], want, 1e-9, "two reeds")
	}

	// One reed cannot beat: the typed value is refused, never swallowed.
	one := NewCurve("clarinet", 1)
	if err := one.SetBeating(60, 2, a4); err == nil {
		t.Fatal("a beating on a one-reed curve must be an error")
	}
	if len(one.Anchors) != 0 {
		t.Fatalf("a refused SetBeating anchored something: %v", one.Anchors)
	}
	if got := one.Beating(60, a4); got != 0 {
		t.Fatalf("one-reed Beating = %v, want 0", got)
	}
}

// A beating writes an ordinary anchor a later Set may edit; a rejected one is a no-op.
func TestSetBeatingWritesAnOrdinaryAnchor(t *testing.T) {
	const a4 = 440
	c := NewCurve("musette", 3)
	c.RefReed = 1
	mustBeat(t, c, 60, 8, a4)
	mustSet(t, c, 60, 2, 5, a4)

	got := c.At(60)
	almost(t, got[0], -8, 1e-12, "the derived reed stands")
	almost(t, got[2], 5, 1e-12, "the hand-edited reed wins")
	if len(c.Anchors) != 1 {
		t.Fatalf("anchors = %d, want 1", len(c.Anchors))
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	if err := c.SetBeating(tuning.MaxNote+1, 2, a4); err == nil {
		t.Fatal("a beating on a note outside the range must be an error")
	}
	c.Unit = UnitHz
	if err := c.SetBeating(tuning.NoteA4, -100000, a4); err == nil {
		t.Fatal("a beating that puts a reed below zero pitch is not a pitch")
	}
	if len(c.Anchors) != 1 || len(c.Anchors[0].Reeds) != 3 {
		t.Fatalf("a rejected SetBeating changed the curve: %+v", c.Anchors)
	}
}
