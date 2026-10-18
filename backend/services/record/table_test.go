package record

import (
	"errors"
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp"
	sessionsvc "smegg.me/smeggtuner/backend/services/session"
)

// The table is the takes joined to the goal curve, in reference units.
func TestTableJoinsTheGoalCurve(t *testing.T) {
	s, sessions := services(t)
	open(t, sessions, 3)
	s.SetArmed(true)
	// A musette goal: reed 1 flat, reed 2 at pitch, reed 3 sharp.
	for reed, dev := range []float64{-8, 0, 8} {
		if err := sessions.SetAnchor(60, reed, dev, "reference"); err != nil {
			t.Fatal(err)
		}
	}

	// Reed 3 is two units sharp of where the curve wants it.
	lock(s, 60, -8, 0, 10)

	table, err := s.Table()
	if err != nil {
		t.Fatal(err)
	}
	row := table.Rows[0]
	if len(row.Reeds) != 3 || row.Slots != 3 {
		t.Fatalf("reeds = %d of %d slots, want 3 of 3", len(row.Reeds), row.Slots)
	}
	third := row.Reeds[2]
	if !near(third.Curr, 10) || !near(third.Goal, 8) || !near(third.Error, 2) {
		t.Fatalf("reed 3 = %+v, want curr 10, goal 8, error 2", third)
	}
	if third.InTol {
		t.Fatal("two units out with a one unit tolerance is not in tune")
	}
	if first := row.Reeds[0]; !first.InTol || !near(first.Error, 0) {
		t.Fatalf("reed 1 = %+v, want it sitting on its goal", first)
	}
	// Beats are the adjacent pairs; 2-3 asks for the goals' 8 and measures the reeds' 10.
	if len(row.Beats) != 2 {
		t.Fatalf("beats = %d, want one per adjacent pair of three reeds", len(row.Beats))
	}
	if b := row.Beats[1]; b.Pair != "2-3" || !near(b.Goal, 8) || !near(b.Curr, 10) {
		t.Fatalf("beat 2-3 = %+v, want goal 8 and curr 10", b)
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

// A slot that did not sound keeps its place: it has no row, its pairs have no beat, and it cannot be edited.
func TestASilentSlotKeepsItsPlace(t *testing.T) {
	s, sessions := services(t)
	open(t, sessions, 3)
	s.SetArmed(true)

	m := measured(48, true, reeds(48, -4)[0], silent(), reeds(48, 4)[0])
	s.OnMeasurement(m)

	row := table(t, s).Rows[0]
	if row.Slots != 3 || len(row.Reeds) != 2 {
		t.Fatalf("row = %d rows of %d slots, want 2 of 3", len(row.Reeds), row.Slots)
	}
	if row.Reeds[0].Reed != 0 || row.Reeds[1].Reed != 2 {
		t.Fatalf("rows are slots %d and %d, want 0 and 2: the third reed must not move into the gap",
			row.Reeds[0].Reed, row.Reeds[1].Reed)
	}
	if len(row.Beats) != 0 {
		t.Fatalf("beats = %+v, want none: both pairs touch the silent slot", row.Beats)
	}
	if _, err := s.EditReed(0, 1, 3, "reference"); !errors.Is(err, sessionsvc.ErrInvalidReed) {
		t.Fatalf("editing the silent slot: err = %v, want %s", err, sessionsvc.ErrInvalidReed.Key)
	}
}

func TestEditReedRecomputesTheRowAndMarksTheTake(t *testing.T) {
	s, sessions := services(t)
	open(t, sessions, 3)
	s.SetArmed(true)
	if err := sessions.SetAnchor(69, 2, 8, "reference"); err != nil {
		t.Fatal(err)
	}
	lock(s, 69, -8, 0, 12)

	table, err := s.EditReed(0, 2, 9, "reference")
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
	if b := row.Beats[1]; b.Pair != "2-3" || !near(b.Curr, 9) {
		t.Fatalf("beat 2-3 = %+v, want it recomputed from the edited reed", b)
	}

	// Hz is converted against the slot's scale pitch at the pass's reference: 2 Hz at A4 is 2/0.264 units.
	table, err = s.EditReed(0, 2, 2, "hz")
	if err != nil {
		t.Fatal(err)
	}
	if got := table.Rows[0].Reeds[2].Curr; !near(got, 2/(a4*0.0006)) {
		t.Fatalf("2 Hz sharp at A4 = %v units, want %v", got, 2/(a4*0.0006))
	}
	// Cents convert through the log form: 12 cents is 2^(12/1200)-1 over 0.0006.
	table, err = s.EditReed(0, 2, 12, "cent")
	if err != nil {
		t.Fatal(err)
	}
	if got := table.Rows[0].Reeds[2].Curr; got < 11.5 || got > 11.6 {
		t.Fatalf("12 cents = %v units, want about 11.56", got)
	}

	if _, err := s.EditReed(9, 0, 1, "reference"); !errors.Is(err, ErrTakeNotFound) {
		t.Fatalf("editing a take that is not there: err = %v, want %s", err, ErrTakeNotFound.Key)
	}
	if _, err := s.EditReed(0, 0, 1, "furlong"); !errors.Is(err, sessionsvc.ErrInvalidUnit) {
		t.Fatalf("an unknown unit: err = %v, want %s", err, sessionsvc.ErrInvalidUnit.Key)
	}
	if _, err := s.EditReed(0, 0, -500, "hz"); !errors.Is(err, sessionsvc.ErrInvalidValue) {
		t.Fatalf("an edit below zero Hz: err = %v, want %s", err, sessionsvc.ErrInvalidValue.Key)
	}
}

// Held and frozen frames repeat a reading the engine already gave; neither is a new take.
func TestHeldAndFrozenFramesAreNotNewReadings(t *testing.T) {
	s, sessions := services(t)
	open(t, sessions, 2)
	s.SetArmed(true)

	s.OnMeasurement(fine(60, true, -4, 4))
	held := dsp.Measurement{State: dsp.StateRunning, Frame: dsp.FrameHeld, Note: 60, Lock: dsp.Lock{Visible: true},
		Reeds: []dsp.ReedMeasure{}}
	for range 3 {
		s.OnMeasurement(held)
	}
	frozen := fine(62, true, -3, 3)
	frozen.State = dsp.StateFrozen
	for range 3 {
		s.OnMeasurement(frozen)
	}
	if got := s.State().Readings; got != 1 {
		t.Fatalf("readings = %d, want 1: a frozen re-emit repeats the take it latched", got)
	}

	// Unfrozen, the engine's own take of the next note is a reading.
	s.OnMeasurement(fine(62, true, -3, 3))
	if got := s.State().Readings; got != 2 {
		t.Fatalf("readings = %d, want 2 once the engine measures the next note itself", got)
	}
}
