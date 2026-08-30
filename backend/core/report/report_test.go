package report

import "testing"

// A slot that did not sound prints absent, and takes both its beats with it.
func TestASilentSlotPrintsAbsent(t *testing.T) {
	r := sheet(t, musette(t))
	m := row(t, r, 64)

	if len(m.Reeds) != 3 {
		t.Fatalf("the row carries %d reed cells, want one per reed the instrument sounds", len(m.Reeds))
	}
	if !m.Reeds[0].Present || m.Reeds[1].Present || !m.Reeds[2].Present {
		t.Fatalf("cells %+v, want the middle one absent and the others read", m.Reeds)
	}
	for i, b := range m.Beats {
		if b.Present {
			t.Errorf("beat column %s printed %+v next to a slot that did not sound", r.Pairs[i].Key, b)
		}
	}
}

func TestRowsCollapseAndMark(t *testing.T) {
	r := sheet(t, musette(t))

	if got := len(r.Rows); got != 5 {
		t.Fatalf("%d rows, want 5 (one per note)", got)
	}

	twice := row(t, r, 62)
	// The last take is the most recent word on the note.
	if got := twice.Reeds[0].Curr; got < -7.7 || got > -7.5 {
		t.Errorf("Curr of reed 1 = %.2f, want the last take (-7.6)", got)
	}

	hand := row(t, r, 67)
	if !hand.Manual {
		t.Error("the hand-edited note is not marked")
	}
	if len(hand.Reeds) != 3 {
		t.Fatalf("a row always carries a cell per reed the instrument sounds, got %d", len(hand.Reeds))
	}
	if hand.Reeds[2].Present {
		t.Error("the reed that never sounded is printed as a reading")
	}
}

// The verdict is the backend's; Error is Curr - Goal against the pass's own reference.
func TestErrorsAgainstTheGoal(t *testing.T) {
	r := sheet(t, musette(t))
	c := row(t, r, 60)

	want := []struct {
		goal, err float64
		inTol     bool
	}{
		{-8, -0.4, true},
		{0, 0.3, true},
		{8, 4.5, false}, // 12.5 measured against a goal of 8
	}
	for i, w := range want {
		got := c.Reeds[i]
		if !near(got.Goal, w.goal, 0.05) {
			t.Errorf("reed %d goal = %.2f, want %.2f", i+1, got.Goal, w.goal)
		}
		if !near(got.Error, w.err, 0.05) {
			t.Errorf("reed %d error = %.2f, want %.2f", i+1, got.Error, w.err)
		}
		if got.InTol != w.inTol {
			t.Errorf("reed %d inTol = %v, want %v", i+1, got.InTol, w.inTol)
		}
	}
	if c.OutOfTol == 0 {
		t.Error("a row with a reed 4.5 units out is not counted as out of tolerance")
	}
}

// No goal curve: Goal is zero, Error is the deviation from the tempered scale.
func TestNoCurveIsOrdinary(t *testing.T) {
	s := musette(t)
	s.Curve = nil

	r := sheet(t, s)
	if r.Identity.HasCurve {
		t.Fatal("a session without a curve claims one")
	}
	c := row(t, r, 60)
	for _, cell := range c.Reeds {
		if cell.Goal != 0 {
			t.Errorf("reed %d goal = %.2f without a curve, want 0", cell.Reed, cell.Goal)
		}
		if cell.Error != cell.Curr {
			t.Errorf("reed %d error = %.2f, want Curr (%.2f) with no goal to subtract",
				cell.Reed, cell.Error, cell.Curr)
		}
	}
}

// The pass's frozen reference is what every number was measured from: the beat converts at its pitch.
func TestPassReferenceIsThePassesOwn(t *testing.T) {
	s := musette(t)
	r := sheet(t, s)

	if r.Session.A4 != s.A4 {
		t.Fatalf("report quotes A4 %.1f, want the session's %.1f", r.Session.A4, s.A4)
	}
	if got := row(t, r, 60).Beats[0].Curr; !near(got, 0.3 - -8.4, 0.01) {
		t.Errorf("beat 1-2 = %.3f units, want 8.70: the beat was read against the wrong reference", got)
	}
}
