package target

import (
	"testing"

	"smegg.me/smeggtuner/backend/core/tuning"
)

// sloped is a three-reed curve whose reeds fall towards the treble: reed 1 negative, reed 2 at
// pitch, reed 3 positive, both outer reeds two units nearer zero per two semitones.
func sloped(t *testing.T) *Curve {
	t.Helper()
	c := NewCurve("sloped", 3)
	for r, v := range []float64{-10, 0, 10} {
		mustSet(t, c, 60, r, v)
	}
	for r, v := range []float64{-8, 0, 8} {
		mustSet(t, c, 62, r, v)
	}
	return c
}

// An extrapolating end continues the line through its two outermost anchors.
func TestExtrapolationContinuesTheLine(t *testing.T) {
	c := sloped(t)
	c.ExtrapolateLeft, c.ExtrapolateRight = true, true

	for n, want := range map[tuning.Note][]float64{
		50: {-20, 0, 20},
		58: {-12, 0, 12},
		65: {-5, 0, 5},
		69: {-1, 0, 1},
	} {
		got := c.At(n)
		for r := range want {
			almost(t, got[r], want[r], 1e-9, "the line through the end anchors")
		}
		if !c.Covers(n) {
			t.Fatalf("note %d is extrapolated and not covered", n)
		}
	}
}

// The continued line stops where it reaches zero; past that the goal is zero, never the other sign.
func TestExtrapolationStopsAtZero(t *testing.T) {
	c := sloped(t)
	c.ExtrapolateRight = true

	for _, n := range []tuning.Note{70, 71, 90, tuning.MaxNote} {
		for r, v := range c.At(n) {
			if v != 0 {
				t.Fatalf("At(%d)[%d] = %v, want zero past the crossing", n, r, v)
			}
		}
		if !c.Covers(n) {
			t.Fatalf("note %d past the crossing is still covered", n)
		}
	}
}

// A line moving away from zero runs to the limit, where the reference drops it.
func TestExtrapolationIsDroppedAboveTheLimit(t *testing.T) {
	c := sloped(t)
	c.ExtrapolateLeft = true

	almost(t, c.At(40)[2], 30, 1e-9, "reed 3 still rising below")
	almost(t, c.At(40)[0], -30, 1e-9, "reed 1 has no lower limit")
	noGoalSlot(t, c, 18, 2, "reed 3 at 52 units, past the 50 limit")
}

// Each end is its own switch, and an end that does not extrapolate sets no goal.
func TestExtrapolationEndsAreApart(t *testing.T) {
	c := sloped(t)
	c.ExtrapolateLeft = true
	almost(t, c.At(58)[2], 12, 1e-9, "left extrapolates")
	noGoal(t, c, 63, "right does not")
	if c.Covers(63) {
		t.Fatal("the right end does not extrapolate and is covered")
	}

	c.ExtrapolateLeft, c.ExtrapolateRight = false, true
	noGoal(t, c, 58, "left does not")
	almost(t, c.At(63)[2], 7, 1e-9, "right extrapolates")
}

func noGoalSlot(t *testing.T, c *Curve, note tuning.Note, slot int, msg string) {
	t.Helper()
	if _, ok := goalAt(c.At(note), slot); ok {
		t.Fatalf("%s: At(%d)[%d] = %v, want no goal", msg, note, slot, c.At(note)[slot])
	}
}
