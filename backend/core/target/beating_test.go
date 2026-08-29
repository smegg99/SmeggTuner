package target

import (
	"fmt"
	"testing"

	"smegg.me/smeggtuner/backend/core/tuning"
)

// hzAtNote reads a stored goal back as the Hz offset it was authored as.
func hzAtNote(t *testing.T, c *Curve, note tuning.Note) []float64 {
	t.Helper()
	goal := c.At(note)
	out := make([]float64, len(goal))
	for i, d := range goal {
		out[i] = UnitHz.FromGoal(d, note)
	}
	return out
}

func mustBeat(t *testing.T, c *Curve, note tuning.Note, value float64) {
	t.Helper()
	if err := c.SetBeating(note, value); err != nil {
		t.Fatalf("SetBeating(%d, %v): %v", note, value, err)
	}
}

// SetBeating derives the minus / zero / plus of a musette from one typed value.
func TestSetBeatingDerivesTheReeds(t *testing.T) {
	c := NewCurve("musette", 3)
	c.Unit = UnitHz
	c.RefReed = 1

	mustBeat(t, c, tuning.NoteA4, 2)
	got := hzAtNote(t, c, tuning.NoteA4)
	for r, want := range []float64{-2, 0, 2} {
		almost(t, got[r], want, 1e-9, "2 Hz beating, reed 2 on zero")
	}
	almost(t, c.Beating(tuning.NoteA4), 2, 1e-9, "Beating reads it straight back")

	k := NewCurve("units", 3)
	k.RefReed = 1
	mustBeat(t, k, 60, 9)
	for r, want := range []float64{-9, 0, 9} {
		almost(t, k.At(60)[r], want, 1e-12, "9 units of beating")
	}
	almost(t, k.Beating(60), 9, 1e-12, "Beating in reference units")

	cents := NewCurve("cents", 3)
	cents.Unit = UnitCents
	cents.RefReed = 1
	mustBeat(t, cents, 60, 9)
	almost(t, cents.Beating(60), 9, 1e-9, "Beating reads typed cents straight back")
	almost(t, UnitCents.FromReference(cents.At(60)[2], 0), 9, 1e-9, "the sharp reed is 9 cents up")
}

// RefReed decides which reed the tremolo hangs off; never assume the middle one.
func TestSetBeatingHonoursRefReed(t *testing.T) {
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
		mustBeat(t, c, tuning.NoteA4, 2)

		got := hzAtNote(t, c, tuning.NoteA4)
		for r := range tc.want {
			almost(t, got[r], tc.want[r], 1e-9, fmt.Sprintf("ref reed %d", tc.ref))
		}
	}
}

// Asymmetry is the old generator's rule: beat 2-3 is k = 1 + asymmetry/100 times beat 1-2, and
// beat 1-2 stays what was typed wherever the reference reed sits.
func TestAsymmetryRaisesTheUpperBeat(t *testing.T) {
	cases := []struct {
		ref  int
		asym float64
		want []float64
	}{
		{1, 0, []float64{-2, 0, 2}},
		{1, 20, []float64{-2, 0, 2.4}},
		{1, -50, []float64{-2, 0, 1}},
		{1, 100, []float64{-2, 0, 4}},
		{1, -100, []float64{-2, 0, 0}},
		{0, 20, []float64{0, 2, 4.4}},
		{2, 20, []float64{-4.4, -2.4, 0}},
	}
	for _, tc := range cases {
		c := NewCurve("musette", 3)
		c.Unit = UnitHz
		c.RefReed = tc.ref
		c.Asymmetry = tc.asym
		mustBeat(t, c, tuning.NoteA4, 2)

		got := hzAtNote(t, c, tuning.NoteA4)
		msg := fmt.Sprintf("ref %d, asymmetry %v", tc.ref, tc.asym)
		for r := range tc.want {
			almost(t, got[r], tc.want[r], 1e-9, msg)
		}
		almost(t, got[1]-got[0], 2, 1e-9, msg+": beat 1-2 is what was typed")
		almost(t, c.Beating(tuning.NoteA4), 2, 1e-9, msg+": Beating reads beat 1-2")
		if err := c.Validate(); err != nil {
			t.Fatalf("%s: %v", msg, err)
		}
	}

	bad := NewCurve("musette", 3)
	bad.Asymmetry = 101
	if err := bad.Validate(); err == nil {
		t.Fatal("asymmetry past 100 percent must not validate")
	}
}

// Past three reeds each pair beats k times the one below it.
func TestSetBeatingAnyReedCount(t *testing.T) {
	five := NewCurve("bass", 5)
	five.Unit = UnitHz
	five.RefReed = 0
	mustBeat(t, five, 84, 2)
	for r, want := range []float64{0, 2, 4, 6, 8} {
		almost(t, hzAtNote(t, five, 84)[r], want, 1e-9, "five reeds, reed 1 on zero")
	}

	five.RefReed = 2
	five.Asymmetry = 50
	mustBeat(t, five, 84, 2)
	for r, want := range []float64{-5, -3, 0, 4.5, 11.25} {
		almost(t, hzAtNote(t, five, 84)[r], want, 1e-9, "five reeds, reed 3 on zero, beats 2, 3, 4.5, 6.75")
	}
	almost(t, five.Beating(84), 2, 1e-9, "five reeds beat at what was typed")

	two := NewCurve("pair", 2)
	two.Unit = UnitHz
	two.RefReed = 0
	mustBeat(t, two, 60, 3)
	for r, want := range []float64{0, 3} {
		almost(t, hzAtNote(t, two, 60)[r], want, 1e-9, "two reeds")
	}

	// One reed cannot beat: the typed value is refused, never swallowed.
	one := NewCurve("clarinet", 1)
	if err := one.SetBeating(60, 2); err == nil {
		t.Fatal("a beating on a one-reed curve must be an error")
	}
	if len(one.Anchors) != 0 {
		t.Fatalf("a refused SetBeating anchored something: %v", one.Anchors)
	}
	if got := one.Beating(60); got != 0 {
		t.Fatalf("one-reed Beating = %v, want 0", got)
	}
}

// A beating writes an ordinary anchor a later Set may edit; a rejected one is a no-op.
func TestSetBeatingWritesAnOrdinaryAnchor(t *testing.T) {
	c := NewCurve("musette", 3)
	c.RefReed = 1
	mustBeat(t, c, 60, 8)
	mustSet(t, c, 60, 2, 5)

	got := c.At(60)
	almost(t, got[0], -8, 1e-12, "the derived reed stands")
	almost(t, got[2], 5, 1e-12, "the hand-edited reed wins")
	if len(c.Anchors) != 1 {
		t.Fatalf("anchors = %d, want 1", len(c.Anchors))
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}

	if err := c.SetBeating(tuning.MaxNote+1, 2); err == nil {
		t.Fatal("a beating on a note outside the range must be an error")
	}
	c.Unit = UnitHz
	if err := c.SetBeating(tuning.NoteA4, -100000); err == nil {
		t.Fatal("a beating that puts a reed below the empty marker is not a goal")
	}
	if len(c.Anchors) != 1 || len(c.Anchors[0].Reeds) != 3 {
		t.Fatalf("a rejected SetBeating changed the curve: %+v", c.Anchors)
	}
	if got := c.Beating(40); got != 0 {
		t.Fatalf("Beating where the curve sets no goal = %v, want 0", got)
	}
}
