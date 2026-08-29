package session

import (
	"testing"
	"time"

	"smegg.me/smeggtuner/backend/core/dsp"
)

// The fit gets the treble rows with every slot in its card column; a bass take is not on the treble curve.
func TestPassReadingsAreLaidOnTheCardColumns(t *testing.T) {
	s := newTestSession(t, 3)
	s.Instrument.Banks = []Bank{BankM1, BankM2, BankM3}
	s.Instrument.Registers = []Register{{Name: "M1M3", Banks: []Bank{BankM1, BankM3}}}
	s.Instrument.BassReeds = 2
	s.Transpose = 2
	s.UpsertTake(Take{Note: 60, Register: "M1M3", Reeds: []dsp.ReedMeasure{{Present: true, Dev: -4}, {Present: true, Dev: 4}}})
	s.UpsertTake(Take{Note: 48, Bass: true, Reeds: reeds(2)})

	got := s.Readings(3)
	if len(got) != 1 {
		t.Fatalf("readings = %d, want the treble take only", len(got))
	}
	r := got[0]
	if len(r.Reeds) != 3 || r.Reeds[0].Dev != -4 || r.Reeds[1].Present || r.Reeds[2].Dev != 4 {
		t.Fatalf("slots %+v, want M1 and M3 in columns 1 and 3 with M2 silent", r.Reeds)
	}
	if r.Transpose != 2 || len(r.Beats) != 0 {
		t.Fatalf("reading %+v, want the session's transposition and no beats", r)
	}
}

// Undo removes the take captured last, which for a replayed voice is not the last in Takes.
func TestUndoRemovesTheNewestCapture(t *testing.T) {
	s := newTestSession(t, 3)
	now := time.Now()
	s.UpsertTake(take(60, 3, now))
	s.UpsertTake(take(62, 3, now.Add(time.Second)))
	s.UpsertTake(take(60, 3, now.Add(2*time.Second)))

	if got := s.Latest(); got != 0 {
		t.Fatalf("latest = %d, want 0: note 60 was replayed in place", got)
	}
	got, ok := s.UndoLast()
	if !ok || got.Note != 60 {
		t.Fatalf("undo = (%d, %v), want the replayed 60", got.Note, ok)
	}
	if len(s.Takes) != 1 || s.Takes[0].Note != 62 {
		t.Fatalf("takes after undo: %+v, want 62 left", s.Takes)
	}
	if empty := newTestSession(t, 3); empty.Latest() != -1 {
		t.Fatal("an empty pass has no latest take")
	}
}
