package session

import (
	"errors"
	"testing"
)

// The transposition rides with the session like A4: set before the first take, then frozen.
func TestTransposeIsStoredAndFrozenByReadings(t *testing.T) {
	s := service(t)
	if err := s.SetTranspose(2); !errors.Is(err, ErrNoSession) {
		t.Fatalf("no session: err = %v, want %s", err, ErrNoSession.Key)
	}
	dto := create(t, s, "Eb accordion", 3, 440)
	if dto.Transpose != 0 {
		t.Fatalf("a new session is transposed by %d", dto.Transpose)
	}

	for _, bad := range []int{-1, 12} {
		if err := s.SetTranspose(bad); !errors.Is(err, ErrInvalidTranspose) {
			t.Fatalf("transpose %d: err = %v, want %s", bad, err, ErrInvalidTranspose.Key)
		}
	}
	if err := s.SetTranspose(3); err != nil {
		t.Fatal(err)
	}
	if got := s.Active().Transpose; got != 3 {
		t.Fatalf("session transpose = %d, want 3", got)
	}
	if got := s.Goal().Transpose; got != 3 {
		t.Fatalf("goal transpose = %d, want 3: the engine has to hear the sounding pitch", got)
	}

	if err := s.UpsertTake(take(69, 440, 0)); err != nil {
		t.Fatal(err)
	}
	if err := s.SetTranspose(0); !errors.Is(err, ErrHasReadings) {
		t.Fatalf("transpose with readings: err = %v, want %s", err, ErrHasReadings.Key)
	}
	d, err := s.Data()
	if err != nil {
		t.Fatal(err)
	}
	if d.Transpose != 3 {
		t.Fatalf("reading data transpose = %d, want the 3 the take was made with", d.Transpose)
	}

	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if back, err := s.Open(dto.ID); err != nil || back.Transpose != 3 {
		t.Fatalf("reopened = %+v (%v), want the transposition kept", back, err)
	}
}
