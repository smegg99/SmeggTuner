// backend/services/session/close_test.go
package session

import (
	"errors"
	"testing"

	"smegg.me/smeggtuner/backend/core/datastore"
)

func TestCloseRetainsSessionWhenPersistenceFails(t *testing.T) {
	s := service(t)
	opened := create(t, s, "Keep this session", 1, 440)
	db := datastore.Get()
	if err := db.Exec("PRAGMA query_only=ON").Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Exec("PRAGMA query_only=OFF").Error })
	if err := s.Close(); !errors.Is(err, ErrSaveFailed) {
		t.Fatalf("Close = %v, want save failure", err)
	}
	if active := s.Active(); active == nil || active.ID != opened.ID {
		t.Fatalf("failed close lost session: %+v", active)
	}
	if err := db.Exec("PRAGMA query_only=OFF").Error; err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("retry close: %v", err)
	}
	if s.Active() != nil {
		t.Fatal("successful close retained session")
	}
}

func TestInstrumentWriteCannotReachAnotherSession(t *testing.T) {
	s := service(t)
	first := create(t, s, "First", 1, 440)
	second := create(t, s, "Second", 1, 440)
	next := second.Instrument
	next.Name = "Stale editor"
	if err := s.SetInstrument(first.ID, next); !errors.Is(err, ErrNoSession) {
		t.Fatalf("stale write = %v", err)
	}
	if got := s.Active(); got.ID != second.ID || got.Instrument.Name == next.Name {
		t.Fatalf("stale editor changed current session: %+v", got)
	}
	next.A4 = 442
	if err := s.SetInstrument(second.ID, next); err != nil {
		t.Fatal(err)
	}
	if got := s.Active(); got.A4 != 442 || got.Instrument.A4 != 442 {
		t.Fatalf("reference was not saved atomically: %+v", got)
	}
	if err := s.UpsertTake(take(69, 442, 0)); err != nil {
		t.Fatal(err)
	}
	next.A4 = 443
	if err := s.SetInstrument(second.ID, next); !errors.Is(err, ErrHasReadings) {
		t.Fatalf("reference changed with readings: %v", err)
	}
	if s.Active().A4 != 442 {
		t.Fatal("refused write changed reference")
	}
}
