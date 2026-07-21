package record

import (
	"errors"
	"testing"

	sessionsvc "smegg.me/smeggtuner/services/session"
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

func TestEachLockIsAReading(t *testing.T) {
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
		t.Fatalf("rows = %d, want three takes from three locks", len(table.Rows))
	}
	if table.Rows[0].Note != 60 || table.Rows[2].Note != 64 {
		t.Fatalf("rows = %+v, want the three notes that were played", table.Rows)
	}
	if s.State().Readings != 3 {
		t.Fatalf("readings = %d, want 3", s.State().Readings)
	}
}

// A heartbeat carries no note or lock; reading one as an unlock would re-record the held note.
func TestHeartbeatsDoNotDuplicateAReading(t *testing.T) {
	s, sessions := services(t)
	open(t, sessions, 3)
	s.SetArmed(true)

	s.OnMeasurement(fine(60, true, -8, 0, 8))
	for i := 0; i < 5; i++ {
		s.OnMeasurement(heartbeat())
		s.OnMeasurement(fine(60, true, -8, 0, 8)) // the note is still being held
	}

	if got := s.State().Readings; got != 1 {
		t.Fatalf("readings = %d, want 1: one note held is one reading", got)
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
	// Undo on an empty pass does nothing.
	if _, err := s.Undo(); err != nil {
		t.Fatalf("undo on an empty pass: %v", err)
	}
	if got := s.State().Readings; got != 0 {
		t.Fatalf("takes = %d, want 0", got)
	}

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
