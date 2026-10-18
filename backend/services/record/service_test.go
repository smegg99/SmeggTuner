package record

import (
	"errors"
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp"
	sessionsvc "smegg.me/smeggtuner/backend/services/session"
)

func TestPlayingWithNoSessionRecordsNothing(t *testing.T) {
	s, _ := services(t)

	lock(s, 60, -8, 0, 8)

	if got := s.State().Readings; got != 0 {
		t.Fatalf("readings = %d with no session open, want 0", got)
	}
	if _, err := s.Table(); !errors.Is(err, sessionsvc.ErrNoSession) {
		t.Fatalf("Table with no session: err = %v, want %s", err, sessionsvc.ErrNoSession.Key)
	}
}

func TestEachTakeIsAReading(t *testing.T) {
	s, sessions := services(t)
	open(t, sessions, 3)
	s.SetArmed(true)

	lock(s, 60, -8, 0, 8)
	lock(s, 62, -7, 0, 9)
	lock(s, 64, -6, 1, 10)

	table, err := s.Table()
	if err != nil {
		t.Fatal(err)
	}
	if len(table.Rows) != 3 {
		t.Fatalf("rows = %d, want three readings from three takes", len(table.Rows))
	}
	if table.Rows[0].Note != 60 || table.Rows[2].Note != 64 {
		t.Fatalf("rows = %+v, want the three notes that were played", table.Rows)
	}
	if s.State().Readings != 3 {
		t.Fatalf("readings = %d, want 3", s.State().Readings)
	}
}

// Only the analysis the engine took is a reading: a held lock, a heartbeat or a lock the engine did not take is not.
func TestOnlyTheEnginesTakeIsAReading(t *testing.T) {
	s, sessions := services(t)
	open(t, sessions, 3)
	s.SetArmed(true)

	seen := capture(t)
	s.OnMeasurement(fine(60, true, -8, 0, 8))
	for i := 0; i < 5; i++ {
		s.OnMeasurement(heartbeat())
		s.OnMeasurement(fine(60, false, -8, 0, 8)) // the note is still held, the engine inhibits it
	}
	locked := fine(62, false, -8, 0, 8)
	locked.Lock = dsp.Lock{Visible: true, Internal: true, Acquired: true}
	s.OnMeasurement(locked)

	if got := s.State().Readings; got != 1 {
		t.Fatalf("readings = %d, want 1: one take is one reading", got)
	}
	tables := 0
	for _, e := range *seen {
		if e.name == EventTable {
			tables++
		}
	}
	if tables != 1 {
		t.Fatalf("published %d tables, want one for the one take", tables)
	}
}

// Chords are never recorded, whatever a frame says.
func TestAChordFrameIsNeverATake(t *testing.T) {
	s, sessions := services(t)
	open(t, sessions, 3)
	s.SetArmed(true)

	chord := fine(60, true, -1, 0, 1)
	chord.Label = "Cm"
	s.OnMeasurement(chord)
	if got := s.State().Readings; got != 0 {
		t.Fatalf("readings = %d after a chord frame, want 0", got)
	}
}

func TestTwoLocksOnOneNoteReplaceIt(t *testing.T) {
	s, sessions := services(t)
	open(t, sessions, 3)
	s.SetArmed(true)

	lock(s, 60, -8, 0, 8)
	lock(s, 60, -6, 2, 10)

	if got := s.State().Readings; got != 1 {
		t.Fatalf("readings = %d, want 1: replaying a note replaces it", got)
	}
	table, err := s.Table()
	if err != nil {
		t.Fatal(err)
	}
	if len(table.Rows) != 1 {
		t.Fatalf("rows = %d, want one", len(table.Rows))
	}
	if got := table.Rows[0].Reeds[0].Curr; got < -6.01 || got > -5.99 {
		t.Fatalf("row shows Curr %v, want the second reading's -6", got)
	}
}

// Undo is Backspace: the last take goes, and with none left it leaves record mode.
func TestUndoAndClear(t *testing.T) {
	s, sessions := services(t)
	open(t, sessions, 3)
	s.SetArmed(true)
	lock(s, 60, 0)
	lock(s, 62, 0)

	table, err := s.Undo()
	if err != nil {
		t.Fatal(err)
	}
	if len(table.Rows) != 1 || table.Rows[0].Note != 60 {
		t.Fatalf("after undo rows = %+v, want the first take only", table.Rows)
	}

	if _, err := s.Undo(); err != nil {
		t.Fatal(err)
	}
	if !s.Armed() {
		t.Fatal("removing the last take keeps record mode on")
	}
	if _, err := s.Undo(); err != nil {
		t.Fatalf("undo on an empty pass: %v", err)
	}
	if got := s.State().Readings; got != 0 || s.Armed() {
		t.Fatalf("takes = %d, armed %v: undo on an empty pass leaves record mode", got, s.Armed())
	}

	s.SetArmed(true)
	lock(s, 60, 0)
	table, err = s.Clear()
	if err != nil {
		t.Fatal(err)
	}
	if len(table.Rows) != 0 {
		t.Fatalf("after clear rows = %d, want none", len(table.Rows))
	}
	if s.State().SessionID == "" {
		t.Fatal("clearing empties the readings, it does not close the session")
	}
}

// The session closing under a running engine is not a fault; the reading goes nowhere.
func TestClosingTheSessionStopsRecording(t *testing.T) {
	s, sessions := services(t)
	open(t, sessions, 3)
	s.SetArmed(true) // otherwise warm-up alone would explain the zero readings below
	if err := sessions.Close(); err != nil {
		t.Fatal(err)
	}

	lock(s, 60, 0)
	if got := s.State().Readings; got != 0 {
		t.Fatalf("readings = %d with the session closed, want 0", got)
	}
}
