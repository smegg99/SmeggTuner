package session

import (
	"errors"
	"strings"
	"testing"
)

// The note is the one thing on the card that is not measured. It is written while
// the work goes on, so unlike the reference pitch it stays open with readings present.
func TestTheNoteIsWrittenWhileTheWorkGoesOn(t *testing.T) {
	s := service(t)
	create(t, s, "Bench", 3, 442)

	const note = "Musette register, low block replaced."
	if err := s.SetNotes(note); err != nil {
		t.Fatalf("SetNotes: %v", err)
	}

	got := s.Active()
	if got == nil {
		t.Fatal("no active session after SetNotes")
	}
	if got.Notes != note {
		t.Errorf("notes = %q, want %q", got.Notes, note)
	}
}

func TestTheNoteSurvivesTheReadingsItDescribes(t *testing.T) {
	s := service(t)
	dto := create(t, s, "Bench", 3, 442)

	if err := s.UpsertTake(take(69, dto.A4, 0)); err != nil {
		t.Fatalf("UpsertTake: %v", err)
	}
	if err := s.SetNotes("written after the first note sounded"); err != nil {
		t.Fatalf("SetNotes with readings: %v", err)
	}
}

func TestTheNoteNeedsAnOpenSession(t *testing.T) {
	s := service(t)

	if err := s.SetNotes("nowhere to put this"); !errors.Is(err, ErrNoSession) {
		t.Errorf("SetNotes with no session = %v, want ErrNoSession", err)
	}
}

// Printed under the table, so an unbounded one runs off the card.
func TestTheNoteIsBounded(t *testing.T) {
	s := service(t)
	create(t, s, "Bench", 3, 442)

	if err := s.SetNotes(strings.Repeat("x", maxNotes)); err != nil {
		t.Errorf("a note of exactly maxNotes was refused: %v", err)
	}
	if err := s.SetNotes(strings.Repeat("x", maxNotes+1)); !errors.Is(err, ErrNotesTooLong) {
		t.Errorf("an over-long note = %v, want ErrNotesTooLong", err)
	}
}

func TestNotesBoundCountsUnicodeCodePoints(t *testing.T) {
	s := service(t)
	create(t, s, "Bench", 3, 442)
	for _, character := range []string{"\u0105", "\U0001f3b5"} {
		text := strings.Repeat(character, maxNotes)
		if err := s.SetNotes(text); err != nil {
			t.Fatalf("500 code points: %v", err)
		}
		if got := s.Active().Notes; got != text {
			t.Fatal("accepted note changed")
		}
		if err := s.SetNotes(text + character); !errors.Is(err, ErrNotesTooLong) {
			t.Fatalf("501 code points: %v", err)
		}
		if got := s.Active().Notes; got != text {
			t.Fatal("rejected note replaced accepted note")
		}
	}
}
