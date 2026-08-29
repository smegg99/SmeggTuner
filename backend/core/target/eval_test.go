package target

import (
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// noGoal fails unless every slot of At(note) is meter.Missing.
func noGoal(t *testing.T, c *Curve, note tuning.Note, msg string) {
	t.Helper()
	for r, v := range c.At(note) {
		if v != meter.Missing {
			t.Fatalf("%s: At(%d)[%d] = %v, want no goal", msg, note, r, v)
		}
	}
}

// The no-goal mode: no anchors gives a pure indicator, out of the same object.
func TestAnEmptyCurveSetsNoGoal(t *testing.T) {
	c := NewCurve("", 3)
	for n := tuning.MinNote; n <= tuning.MaxNote; n++ {
		if got := c.At(n); len(got) != 3 {
			t.Fatalf("At(%d) gave %d values, want 3", n, len(got))
		}
		noGoal(t, c, n, "empty curve")
		if c.Covers(n) {
			t.Fatalf("an empty curve covers note %d", n)
		}
	}
}

func TestNilCurveIsSafe(t *testing.T) {
	var c *Curve
	if got := c.At(tuning.NoteA4); len(got) != 0 {
		t.Fatalf("nil curve At = %v, want empty", got)
	}
	if c.Covers(tuning.NoteA4) {
		t.Fatal("a nil curve covers nothing")
	}
	if err := c.Validate(); err != nil {
		t.Fatalf("nil curve Validate: %v", err)
	}
}

// One anchor is one note of goal: there is no line to continue, even with both ends extrapolating.
func TestOneAnchorCoversItsNoteOnly(t *testing.T) {
	c := NewCurve("one", 3)
	c.ExtrapolateLeft, c.ExtrapolateRight = true, true
	mustSet(t, c, tuning.NoteA4, 0, -8)
	mustSet(t, c, tuning.NoteA4, 2, 8)

	for r, want := range []float64{-8, 0, 8} {
		almost(t, c.At(tuning.NoteA4)[r], want, 1e-12, "the anchor")
	}
	for _, n := range []tuning.Note{tuning.MinNote, tuning.NoteA4 - 1, tuning.NoteA4 + 1, tuning.MaxNote} {
		noGoal(t, c, n, "one anchor")
		if c.Covers(n) {
			t.Fatalf("one anchor covers note %d", n)
		}
	}
}

func TestTwoAnchorsInterpolate(t *testing.T) {
	c := NewCurve("two", 3)
	mustSet(t, c, 60, 2, 4)
	mustSet(t, c, 72, 2, 16)

	almost(t, c.At(66)[2], 10, 1e-12, "midpoint")
	almost(t, c.At(63)[2], 7, 1e-12, "quarter")
	almost(t, c.At(60)[2], 4, 1e-12, "on the low anchor")
	almost(t, c.At(72)[2], 16, 1e-12, "on the high anchor")
	almost(t, c.At(61)[0], 0, 1e-12, "an untouched reed of an anchor is at pitch")
	for _, n := range []tuning.Note{60, 66, 72} {
		if !c.Covers(n) {
			t.Fatalf("note %d is between the anchors", n)
		}
	}
}

// The reference beat list has no extrapolation: past the outermost anchors there is no goal, not zero.
func TestNoGoalPastTheAnchors(t *testing.T) {
	c := NewCurve("musette", 3)
	mustSet(t, c, 60, 2, 4)
	mustSet(t, c, 72, 2, 16)

	for _, n := range []tuning.Note{tuning.MinNote, 59, 73, tuning.MaxNote} {
		noGoal(t, c, n, "outside the anchors")
		if c.Covers(n) {
			t.Fatalf("note %d is outside the anchors and covered", n)
		}
	}
	noGoal(t, c, 0, "below the range")
	noGoal(t, c, 200, "above the range")
}

// A filled value above 50 units is dropped, after interpolation: the line has a gap there.
func TestValuesAboveFiftyAreDropped(t *testing.T) {
	c := NewCurve("wide", 1)
	mustSet(t, c, 60, 0, 0)
	mustSet(t, c, 72, 0, 60)

	almost(t, c.At(70)[0], 50, 1e-9, "50 itself is a goal")
	noGoal(t, c, 71, "55 units")
	noGoal(t, c, 72, "an anchor above 50")
	if !c.Covers(71) {
		t.Fatal("a dropped value still lies inside the anchors")
	}
}

// Hz is read at the note's pitch for A4 = 440 Hz with the reference's widened 0.06, whatever A4 is.
func TestHzGoalsReadAgainstA440(t *testing.T) {
	c := NewCurve("hz", 3)
	c.Unit = UnitHz
	mustSet(t, c, tuning.NoteA4, 1, 2)
	almost(t, c.At(tuning.NoteA4)[1], 2*100/(440*meter.TargetHertzScale), 1e-12, "2 Hz at A4")
	mustSet(t, c, tuning.NoteA4+12, 1, 2)
	almost(t, c.At(tuning.NoteA4 + 12)[1], 2*100/(880*meter.TargetHertzScale), 1e-12, "2 Hz an octave up is half the units")

	for _, tc := range []struct {
		note tuning.Note
		hz   float64
	}{{tuning.NoteA4, -2}, {tuning.MinNote, 0.35}, {60, 1.25}, {tuning.MaxNote, 40}} {
		d := UnitHz.ToGoal(tc.hz, tc.note)
		almost(t, UnitHz.FromGoal(d, tc.note), tc.hz, 1e-12, "hz round trip")
	}
}

// Cents in, reference units stored, cents back out.
func TestCentsAuthoredRoundTrips(t *testing.T) {
	c := NewCurve("cents", 1)
	c.Unit = UnitCents
	for _, cents := range []float64{-30, -1, 0, 1.5, 12, 50} {
		mustSet(t, c, 60, 0, cents)
		almost(t, UnitCents.FromReference(c.At(60)[0], 0), cents, 1e-9, "cents round trip")
	}
	mustSet(t, c, 60, 0, 1200)
	almost(t, c.Anchors[0].Reeds[0], 1/0.0006, 1e-6, "an octave up is the note's frequency again in units")
	noGoal(t, c, 60, "an octave is far above the 50 unit limit")
}

// An anchor typed for reed 1 alone leaves the others at zero.
func TestPartialAnchor(t *testing.T) {
	c := &Curve{ReedCount: 3, RefReed: 1, Unit: UnitReference, Interpolate: true,
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
	mustSet(t, c, 60, 0, 4)
	mustSet(t, c, 72, 0, 16)

	almost(t, c.At(66)[0], 10, 1e-12, "on by default: a ramp")

	c.Interpolate = false
	almost(t, c.At(60)[0], 4, 1e-12, "an anchor is always itself")
	almost(t, c.At(72)[0], 16, 1e-12, "an anchor is always itself")
	almost(t, c.At(62)[0], 4, 1e-12, "nearer the low anchor")
	almost(t, c.At(70)[0], 16, 1e-12, "nearer the high anchor")
	almost(t, c.At(66)[0], 4, 1e-12, "exactly between: the value it is already holding")
	almost(t, c.At(67)[0], 16, 1e-12, "past the middle it steps")
	noGoal(t, c, 40, "still nothing below")
	noGoal(t, c, 100, "still nothing above")
}
