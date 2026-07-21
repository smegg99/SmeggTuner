package target

import (
	"testing"

	"smegg.me/smeggtuner/core/tuning"
)

// The no-goal mode: no target gives a pure indicator, out of the same object.
func TestEmptyCurveIsZeros(t *testing.T) {
	c := NewCurve("", 3)
	for n := tuning.MinNote; n <= tuning.MaxNote; n++ {
		got := c.At(n)
		if len(got) != 3 {
			t.Fatalf("At(%d) gave %d values, want 3", n, len(got))
		}
		for r, v := range got {
			if v != 0 {
				t.Fatalf("At(%d)[%d] = %v, want 0", n, r, v)
			}
		}
	}
}

func TestNilCurveIsSafe(t *testing.T) {
	var c *Curve
	if got := c.At(tuning.NoteA4); len(got) != 0 {
		t.Fatalf("nil curve At = %v, want empty", got)
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("nil curve Validate: %v", err)
	}
}

func TestOneAnchorExtrapolatesFlat(t *testing.T) {
	c := NewCurve("one", 3)
	mustSet(t, c, tuning.NoteA4, 0, -8, 440)
	mustSet(t, c, tuning.NoteA4, 2, 8, 440)

	want := []float64{-8, 0, 8}
	for _, n := range []tuning.Note{tuning.MinNote, 40, tuning.NoteA4, 100, tuning.MaxNote} {
		got := c.At(n)
		for r := range want {
			almost(t, got[r], want[r], 1e-12, "flat both ways")
		}
	}
}

func TestTwoAnchorsInterpolate(t *testing.T) {
	c := NewCurve("two", 3)
	mustSet(t, c, 60, 2, 4, 440)
	mustSet(t, c, 72, 2, 16, 440)

	almost(t, c.At(66)[2], 10, 1e-12, "midpoint")
	almost(t, c.At(63)[2], 7, 1e-12, "quarter")
	almost(t, c.At(60)[2], 4, 1e-12, "on the low anchor")
	almost(t, c.At(72)[2], 16, 1e-12, "on the high anchor")
	almost(t, c.At(61)[0], 0, 1e-12, "an untouched reed stays at zero")
}

// Linear extrapolation would put C8 at +64 cents; flat is the only defensible
// answer past the last anchor.
func TestExtrapolationIsFlatNotLinear(t *testing.T) {
	c := NewCurve("musette", 3)
	mustSet(t, c, 60, 2, 4, 440)
	mustSet(t, c, 72, 2, 16, 440)

	almost(t, c.At(96)[2], 16, 1e-12, "two octaves above the last anchor")
	almost(t, c.At(tuning.MaxNote)[2], 16, 1e-12, "top of the range")
	almost(t, c.At(36)[2], 4, 1e-12, "two octaves below the first")
	almost(t, c.At(tuning.MinNote)[2], 4, 1e-12, "bottom of the range")
}

// Hz in, cents stored, Hz back out: the display unit may flip at any time, so
// the trip has to be lossless.
func TestHzAuthoredRoundTrips(t *testing.T) {
	c := NewCurve("hz", 3)
	c.Unit = UnitHz
	const a4 = 442

	cases := []struct {
		note tuning.Note
		hz   float64
	}{
		{tuning.NoteA4, 2.0},
		{tuning.NoteA4, -2.0},
		{tuning.MinNote, 0.35},
		{tuning.MaxNote, 40},
		{60, 1.25},
	}
	for _, tc := range cases {
		mustSet(t, c, tc.note, 1, tc.hz, a4)
		cents := c.At(tc.note)[1]
		almost(t, HzFromCents(tc.note, cents, a4), tc.hz, 1e-9, "hz round trip")
	}

	// The stored value really is cents, not the Hz typed: +2 Hz on A4 at 442.
	mustSet(t, c, tuning.NoteA4, 1, 2.0, a4)
	almost(t, c.At(tuning.NoteA4)[1], tuning.Cents(444, 442), 1e-12, "stored as cents")

	oct := HzFromCents(tuning.NoteA4+12, c.At(tuning.NoteA4)[1], a4)
	almost(t, oct, 4.0, 1e-9, "2 Hz at A4 is 4 Hz at A5")
}

// An anchor typed for reed 1 alone leaves the others at zero.
func TestPartialAnchor(t *testing.T) {
	// Flags spelled out because this is a literal, not a NewCurve: they default
	// to true, and a zero-value bool would switch interpolation off.
	c := &Curve{ReedCount: 3, RefReed: 1, Unit: UnitCents,
		Interpolate: true, ExtrapolateLeft: true, ExtrapolateRight: true,
		Anchors: []Anchor{{Note: 60, Reeds: []float64{-5}}, {Note: 72, Reeds: []float64{-9, 0, 9}}}}
	if err := c.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	got := c.At(66)
	almost(t, got[0], -7, 1e-12, "reed 1 interpolates")
	almost(t, got[1], 0, 1e-12, "reed 2 is at pitch")
	almost(t, got[2], 4.5, 1e-12, "reed 3 rises from an implied zero")
}

// Interpolate off: a note between two anchors takes the nearer anchor's value.
func TestInterpolateOff(t *testing.T) {
	c := NewCurve("steps", 1)
	mustSet(t, c, 60, 0, 4, 440)
	mustSet(t, c, 72, 0, 16, 440)

	almost(t, c.At(66)[0], 10, 1e-12, "on by default: a ramp")

	c.Interpolate = false
	almost(t, c.At(60)[0], 4, 1e-12, "an anchor is always itself")
	almost(t, c.At(72)[0], 16, 1e-12, "an anchor is always itself")
	almost(t, c.At(62)[0], 4, 1e-12, "nearer the low anchor")
	almost(t, c.At(70)[0], 16, 1e-12, "nearer the high anchor")
	almost(t, c.At(66)[0], 4, 1e-12, "exactly between: the value it is already holding")
	almost(t, c.At(67)[0], 16, 1e-12, "past the middle it steps")
	almost(t, c.At(40)[0], 4, 1e-12, "still flat below")
	almost(t, c.At(100)[0], 16, 1e-12, "still flat above")
}

// An extrapolate flag off: the curve says nothing past that end.
func TestExtrapolateOff(t *testing.T) {
	c := NewCurve("ends", 2)
	mustSet(t, c, 60, 1, 4, 440)
	mustSet(t, c, 72, 1, 16, 440)

	c.ExtrapolateLeft = false
	almost(t, c.At(40)[1], 0, 1e-12, "below the first anchor: no goal")
	almost(t, c.At(60)[1], 4, 1e-12, "the anchor itself is not extrapolation")
	almost(t, c.At(66)[1], 10, 1e-12, "between the anchors is not extrapolation")
	almost(t, c.At(100)[1], 16, 1e-12, "the other end is still held flat")

	c.ExtrapolateLeft = true
	c.ExtrapolateRight = false
	almost(t, c.At(40)[1], 4, 1e-12, "flat again below")
	almost(t, c.At(72)[1], 16, 1e-12, "the anchor itself is not extrapolation")
	almost(t, c.At(100)[1], 0, 1e-12, "above the last anchor: no goal")

	one := NewCurve("single", 1)
	one.ExtrapolateLeft, one.ExtrapolateRight = false, false
	mustSet(t, one, 60, 0, 7, 440)
	almost(t, one.At(60)[0], 7, 1e-12, "the anchor")
	almost(t, one.At(59)[0], 0, 1e-12, "a semitone below it")
	almost(t, one.At(61)[0], 0, 1e-12, "a semitone above it")
}
