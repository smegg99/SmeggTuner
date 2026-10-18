package session

import (
	"errors"
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// A4 is frozen once there are readings: each is a deviation from the A4 it was taken at.
func TestA4IsFrozenOnceThereAreReadings(t *testing.T) {
	s := service(t)
	create(t, s, "Morino", 3, 440)

	if err := s.SetA4(442); err != nil {
		t.Fatalf("SetA4 with nothing recorded: %v", err)
	}
	if err := s.UpsertTake(take(69, 442, -3, 0, 3)); err != nil {
		t.Fatal(err)
	}
	if err := s.SetA4(445); !errors.Is(err, ErrHasReadings) {
		t.Fatalf("SetA4 with readings = %v, want ErrHasReadings", err)
	}
}

func TestReadingsLifecycle(t *testing.T) {
	s := service(t)
	create(t, s, "Morino", 3, 440)

	if err := s.UpsertTake(take(69, 440, 0)); err != nil {
		t.Fatal(err)
	}
	// The same voice again replaces it.
	if err := s.UpsertTake(take(69, 440, 5)); err != nil {
		t.Fatal(err)
	}
	if got := len(s.Snapshot().Takes); got != 1 {
		t.Fatalf("readings = %d, want 1 after replaying the same note", got)
	}
	if _, err := s.UndoTake(); err != nil {
		t.Fatal(err)
	}
	if got := len(s.Snapshot().Takes); got != 0 {
		t.Fatalf("readings = %d after undo, want 0", got)
	}
}

// Everything recorded must reach the disk; a tuning history cannot be recomputed.
func TestTakesReachTheDisk(t *testing.T) {
	s := service(t)

	dto := create(t, s, "Morino", 3, 440)
	for _, n := range []tuning.Note{60, 62, 64} {
		if err := s.UpsertTake(take(n, 440, -8, 0, 8)); err != nil {
			t.Fatal(err)
		}
	}

	// Closing writes it out; a fresh service over the same datastore simulates a restart.
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	next := New()
	t.Cleanup(func() { _ = next.ServiceShutdown() })
	reopened, err := next.Open(dto.ID)
	if err != nil {
		t.Fatal(err)
	}
	if reopened.Readings != 3 {
		t.Fatalf("reopened readings = %d, want the three that reached the disk", reopened.Readings)
	}
}

func TestDeleteClosesTheActiveSession(t *testing.T) {
	s := service(t)
	dto := create(t, s, "Morino", 3, 440)

	if err := s.Delete(dto.ID); err != nil {
		t.Fatal(err)
	}
	if s.Active() != nil {
		t.Fatal("deleting the open session closes it")
	}
	list, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("List() = %d, want none", len(list))
	}
}

// A hand edit lands in reference units on a sounding slot, and a slot that never sounded refuses it.
func TestSetTakeReedEditsOnlyASoundingSlot(t *testing.T) {
	s := service(t)
	create(t, s, "Morino", 3, 440)

	tk := take(69, 440, -8, 0, 8)
	tk.Reeds[1] = dsp.ReedMeasure{Dev: meter.Missing}
	if err := s.UpsertTake(tk); err != nil {
		t.Fatal(err)
	}
	if err := s.SetTakeReed(0, 2, 5); err != nil {
		t.Fatal(err)
	}
	got := s.Snapshot().Takes[0]
	if got.Reeds[2].Dev != 5 || !got.Manual {
		t.Fatalf("edited take = %+v, want slot 3 at 5 units and the take marked manual", got)
	}
	if err := s.SetTakeReed(0, 1, 3); !errors.Is(err, ErrInvalidReed) {
		t.Fatalf("editing a slot that did not sound: err = %v, want %s", err, ErrInvalidReed.Key)
	}
	if err := s.SetTakeReed(4, 0, 3); !errors.Is(err, ErrTakeNotFound) {
		t.Fatalf("editing a take that is not there: err = %v, want %s", err, ErrTakeNotFound.Key)
	}
}
