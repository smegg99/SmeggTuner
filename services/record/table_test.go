package record

import (
	"errors"
	"testing"

	sessionsvc "smegg.me/smeggtuner/services/session"
)

// The table is the takes joined to the goal curve.
func TestTableJoinsTheGoalCurve(t *testing.T) {
	s, sessions := services(t)
	open(t, sessions, 3)
	s.SetArmed(true)
	// A musette goal: reed 1 flat, reed 2 at pitch, reed 3 sharp.
	for reed, cents := range []float64{-8, 0, 8} {
		if err := sessions.SetAnchor(60, reed, cents, "cent"); err != nil {
			t.Fatal(err)
		}
	}

	// Reed 3 is two cents sharp of where the curve wants it.
	lock(s, 60, -8, 0, 10)

	table, err := s.Table()
	if err != nil {
		t.Fatal(err)
	}
	row := table.Rows[0]
	if len(row.Reeds) != 3 {
		t.Fatalf("reeds = %d, want 3", len(row.Reeds))
	}
	third := row.Reeds[2]
	if !near(third.Curr, 10) || !near(third.Goal, 8) || !near(third.Error, 2) {
		t.Fatalf("reed 3 = %+v, want curr 10, goal 8, error 2", third)
	}
	if third.InTol {
		t.Fatal("two cents out with a one cent tolerance is not in tune")
	}
	if first := row.Reeds[0]; !first.InTol || !near(first.Error, 0) {
		t.Fatalf("reed 1 = %+v, want it sitting on its goal", first)
	}
	// The beats fall out of the reeds' goals: -8 and +8 is a 16 cent tremolo.
	if len(row.Beats) != 3 {
		t.Fatalf("beats = %d, want one per pair of three reeds", len(row.Beats))
	}
	if b := row.Beats[1]; b.Pair != "1-3" || !near(b.Goal, 16) || !near(b.Curr, 18) {
		t.Fatalf("beat 1-3 = %+v, want goal 16 and curr 18", b)
	}
	if table.A4 != a4 || table.ReedCount != 3 {
		t.Fatalf("table = a4 %v / reeds %d, want %v / 3", table.A4, table.ReedCount, a4)
	}
}

// With no curve the goal is zero and the error is the plain deviation from the scale.
func TestTableWithNoCurveIsAPureIndicator(t *testing.T) {
	s, sessions := services(t)
	open(t, sessions, 3)
	s.SetArmed(true)
	lock(s, 60, -8, 0, 8)

	table, err := s.Table()
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range table.Rows[0].Reeds {
		if r.Goal != 0 || !near(r.Error, r.Curr) {
			t.Fatalf("reed %+v, want goal 0 and error == curr with no curve", r)
		}
	}
}

// A merged take carries the flag to the table; core/target still computes per-reed numbers, and the flag is what stops the UI printing them.
func TestMergedReedsAreFlaggedAndCannotBeEditedPerReed(t *testing.T) {
	s, sessions := services(t)
	open(t, sessions, 3)
	s.SetArmed(true)

	s.OnMeasurement(merged(48, false))
	s.OnMeasurement(merged(48, true))

	table, err := s.Table()
	if err != nil {
		t.Fatal(err)
	}
	row := table.Rows[0]
	if !row.ReedsMerged {
		t.Fatal("a merged pair must reach the table flagged: the per-reed rows are not reeds")
	}
	// The beat is the only reading of such a note.
	if len(row.Beats) != 1 || !row.Beats[0].FromEnvelope {
		t.Fatalf("beats = %+v, want the envelope beat the engine measured", row.Beats)
	}

	if _, err := s.EditReed(0, 0, 3, "cent"); !errors.Is(err, ErrReedsMerged) {
		t.Fatalf("editing a reed of a merged pair: err = %v, want %s", err, ErrReedsMerged.Key)
	}
}

// A pair recovered from its beat was measured, so the table carries both flags and the row edits like any other.
func TestReedsRecoveredFromTheBeatAreReeds(t *testing.T) {
	s, sessions := services(t)
	open(t, sessions, 2)
	s.SetArmed(true)

	s.OnMeasurement(recovered(48, false))
	s.OnMeasurement(recovered(48, true))

	table, err := s.Table()
	if err != nil {
		t.Fatal(err)
	}
	row := table.Rows[0]
	if !row.ReedsMerged || !row.ReedsFromBeat {
		t.Fatalf("row = merged %v / fromBeat %v, want both: the spectrum failed and the beat did not",
			row.ReedsMerged, row.ReedsFromBeat)
	}
	if len(row.Reeds) != 2 {
		t.Fatalf("reeds = %d, want the two the beat recovered", len(row.Reeds))
	}
	if _, err := s.EditReed(0, 0, -5, "cent"); err != nil {
		t.Fatalf("editing a recovered reed: %v", err)
	}
}

func TestEditReedRecomputesTheRowAndMarksTheTake(t *testing.T) {
	s, sessions := services(t)
	open(t, sessions, 3)
	s.SetArmed(true)
	if err := sessions.SetAnchor(69, 2, 8, "cent"); err != nil {
		t.Fatal(err)
	}
	lock(s, 69, -8, 0, 12)

	table, err := s.EditReed(0, 2, 9, "cent")
	if err != nil {
		t.Fatal(err)
	}
	row := table.Rows[0]
	if !row.Manual {
		t.Fatal("a hand-edited take must say it was hand-edited: a report has to be able to")
	}
	if !near(row.Reeds[2].Curr, 9) || !near(row.Reeds[2].Error, 1) {
		t.Fatalf("edited reed = %+v, want curr 9 against a goal of 8", row.Reeds[2])
	}
	// The beat between reeds 2 and 3 is derived from them, so it follows the edit.
	if b := row.Beats[2]; b.Pair != "2-3" || !near(b.Curr, 9) {
		t.Fatalf("beat 2-3 = %+v, want it recomputed from the edited reed", b)
	}

	// Hz is converted against the note's scale pitch at the pass's reference.
	table, err = s.EditReed(0, 2, 2, "hz")
	if err != nil {
		t.Fatal(err)
	}
	if got := table.Rows[0].Reeds[2].Curr; got < 7.5 || got > 8.5 {
		t.Fatalf("2 Hz sharp at A4 = %v cents, want about 7.9", got)
	}

	if _, err := s.EditReed(9, 0, 1, "cent"); !errors.Is(err, ErrTakeNotFound) {
		t.Fatalf("editing a take that is not there: err = %v, want %s", err, ErrTakeNotFound.Key)
	}
	if _, err := s.EditReed(0, 0, 1, "furlong"); !errors.Is(err, sessionsvc.ErrInvalidUnit) {
		t.Fatalf("an unknown unit: err = %v, want %s", err, sessionsvc.ErrInvalidUnit.Key)
	}
}
