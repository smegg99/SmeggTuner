// backend/services/session/persistence_test.go
package session

import (
	"errors"
	"sync"
	"testing"

	"smegg.me/smeggtuner/backend/core/datastore"
	"smegg.me/smeggtuner/backend/core/repositories"
	coresession "smegg.me/smeggtuner/backend/core/session"
)

func TestUsedRegistersCannotReinterpretReadings(t *testing.T) {
	for _, kind := range []string{"rename", "remove", "replace banks", "bass ranks"} {
		t.Run(kind, func(t *testing.T) {
			s := openMusette(t)
			if kind == "bass ranks" {
				inst := musette()
				inst.BassReeds = 4
				inst.BassRegisters = []coresession.BassRegister{{Name: "Bass", Feet: []int{16, 8}}}
				if err := s.SetInstrument(s.Active().ID, inst); err != nil {
					t.Fatal(err)
				}
				if err := s.SetBass(true); err != nil {
					t.Fatal(err)
				}
				if err := s.SetBassRegister("Bass"); err != nil {
					t.Fatal(err)
				}
			} else if err := s.SetRegister("M2"); err != nil {
				t.Fatal(err)
			}
			if err := s.UpsertTake(take(69, 440, 0)); err != nil {
				t.Fatal(err)
			}
			before := s.Snapshot()
			next := cloneInstrument(before.Instrument)
			switch kind {
			case "rename":
				next.Registers[2].Name = "Changed"
			case "remove":
				next.Registers = nil
			case "replace banks":
				next.Registers[2].Banks = []coresession.Bank{coresession.BankM3}
			case "bass ranks":
				next.BassRegisters[0].Feet = []int{8, 4}
			}
			if err := s.SetInstrument(before.ID, next); !errors.Is(err, ErrHasReadings) {
				t.Fatalf("edit=%v, want ErrHasReadings", err)
			}
			if err := s.Save(); err != nil {
				t.Fatal(err)
			}
			saved, err := s.sessions().Get(before.ID)
			if err != nil {
				t.Fatal(err)
			}
			if takesDiffer(before, saved.Instrument) {
				t.Fatal("rejected edit reached storage")
			}
		})
	}
}

func TestAdoptionConcurrentWithReadingsPreservesLatestSession(t *testing.T) {
	s := openMusette(t)
	id := s.Active().ID
	var wg sync.WaitGroup
	errs := make(chan error, 2)
	wg.Add(2)
	go func() { defer wg.Done(); _, err := s.AdoptInstrument(id, "remembered"); errs <- err }()
	go func() {
		defer wg.Done()
		for range 25 {
			if err := s.UpsertTake(take(69, 440, 3)); err != nil {
				errs <- err
				return
			}
			if err := s.Save(); err != nil {
				errs <- err
				return
			}
		}
		errs <- nil
	}()
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	got, err := s.sessions().Get(id)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Takes) != 1 || got.Takes[0].Reeds[0].Dev != 3 || got.InstrumentID == "" {
		t.Fatalf("lost reading/link: %+v", got)
	}
}

func TestInstrumentLinkDoesNotOverwriteOrResurrectSession(t *testing.T) {
	s := openMusette(t)
	id := s.Active().ID
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if err := datastore.Get().Model(&coresession.Session{}).Where("id = ?", id).Update("notes", "newer write").Error; err != nil {
		t.Fatal(err)
	}
	if err := s.sessions().LinkInstrument(id, "link"); err != nil {
		t.Fatal(err)
	}
	got, err := s.sessions().Get(id)
	if err != nil {
		t.Fatal(err)
	}
	if got.Notes != "newer write" || got.InstrumentID != "link" {
		t.Fatalf("link overwrote session: %+v", got)
	}
	if err := s.sessions().Delete(id); err != nil {
		t.Fatal(err)
	}
	if err := s.sessions().LinkInstrument(id, "other"); !errors.Is(err, repositories.ErrNotFound) {
		t.Fatalf("missing update=%v", err)
	}
	if _, err := s.sessions().Get(id); !errors.Is(err, repositories.ErrNotFound) {
		t.Fatalf("deleted session resurrected: %v", err)
	}
}

func TestOpeningCurrentSessionKeepsUnflushedChanges(t *testing.T) {
	s := openMusette(t)
	id := s.Active().ID
	s.mu.Lock()
	s.active.Notes = "not flushed yet"
	s.mu.Unlock()
	opened, err := s.Open(id)
	if err != nil {
		t.Fatal(err)
	}
	if opened.Notes != "not flushed yet" {
		t.Fatalf("reopen lost in-memory changes: %+v", opened)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	saved, err := s.sessions().Get(id)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Notes != "not flushed yet" {
		t.Fatalf("close persisted stale snapshot: %+v", saved)
	}
}
