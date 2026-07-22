package session

import (
	"errors"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/datastore/datastoretest"
	"smegg.me/smeggtuner/core/dsp"
	coresession "smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/tuning"
)

func service(t *testing.T) *Service {
	t.Helper()
	datastoretest.Init(t)
	s := New()
	t.Cleanup(func() { _ = s.ServiceShutdown() })
	return s
}

func instrument(reeds int) coresession.Instrument {
	return coresession.Instrument{Make: "Hohner", Model: "Morino", ReedCount: reeds}
}

func create(t *testing.T, s *Service, name string, reeds int, a4 float64) *SessionDTO {
	t.Helper()
	inst := instrument(reeds)
	inst.A4 = a4
	dto, err := s.Create(NewSessionDTO{Name: name, Instrument: inst})
	if err != nil {
		t.Fatalf("create %q: %v", name, err)
	}
	return dto
}

func take(note tuning.Note, a4 float64, cents ...float64) coresession.Take {
	ref := note.Freq(a4)
	m := dsp.Measurement{Note: note, ScalePitch: ref, ReedsSeparated: true}
	for _, c := range cents {
		m.Reeds = append(m.Reeds, dsp.ReedMeasure{Freq: tuning.FreqAtCents(ref, c), DevCents: c})
	}
	return coresession.TakeFrom(m, time.Now())
}

// No session is a legal state, not an error; the reported goal is the empty curve.
func TestNoSessionIsALegalState(t *testing.T) {
	s := service(t)

	if a := s.Active(); a != nil {
		t.Fatalf("Active() = %+v, want nil with nothing open", a)
	}
	if g := s.Goal(); g.Curve != nil || g.A4 != 0 || g.Reeds != 0 || g.Banks != nil {
		t.Fatalf("Goal() = %+v, want the zero goal with nothing open", g)
	}
	list, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("List() = %d sessions, want none: nothing may be created behind the user's back", len(list))
	}
	if err := s.Save(); err != nil {
		t.Fatalf("Save with nothing open must be a no-op: %v", err)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("Close with nothing open must be a no-op: %v", err)
	}
}

func TestCreateOpensAndPersists(t *testing.T) {
	s := service(t)
	dto := create(t, s, "Morino", 3, 442)

	if s.Active() == nil || s.Active().ID != dto.ID {
		t.Fatal("Create must open what it made: the technician said he is tuning this one")
	}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	list, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != dto.ID || list[0].A4 != 442 {
		t.Fatalf("List() = %+v, want the session that was just created", list)
	}

	reopened, err := s.Open(dto.ID)
	if err != nil {
		t.Fatal(err)
	}
	if reopened.A4 != 442 || reopened.Instrument.ReedCount != 3 {
		t.Fatalf("reopened = %+v, want a4 442 and 3 reeds", reopened)
	}
}

func TestCreateValidates(t *testing.T) {
	s := service(t)

	var se *ServiceError
	_, err := s.Create(NewSessionDTO{Name: "", Instrument: instrument(3)})
	if !errors.As(err, &se) || se.Key != ErrInvalidName.Key {
		t.Fatalf("empty name: err = %v, want %s", err, ErrInvalidName.Key)
	}
	// A4 rides on the instrument now, so a bad reference is a bad instrument.
	bad := instrument(3)
	bad.A4 = 400
	_, err = s.Create(NewSessionDTO{Name: "x", Instrument: bad})
	if !errors.Is(err, ErrInvalidA4) {
		t.Fatalf("a4 400: err = %v, want %s", err, ErrInvalidA4.Key)
	}
	_, err = s.Create(NewSessionDTO{Name: "x", Instrument: instrument(9)})
	if !errors.Is(err, ErrInvalidReedCount) {
		t.Fatalf("9 reeds: err = %v, want %s", err, ErrInvalidReedCount.Key)
	}
}

// A five-reed instrument keeps five reeds and a five-wide curve; the engine clamp is services/tuner's.
func TestFiveReedSessionOpensAndKeepsItsReeds(t *testing.T) {
	s := service(t)
	dto := create(t, s, "Bass", 5, 440)

	if dto.Instrument.ReedCount != 5 {
		t.Fatalf("instrument reeds = %d, want the 5 it was created with", dto.Instrument.ReedCount)
	}
	if g := s.Goal(); g.Reeds != 5 {
		t.Fatalf("Goal().Reeds = %d, want 5: the instrument is described as what it is", g.Reeds)
	}
	if err := s.SetAnchor(60, 4, 12, "cent"); err != nil {
		t.Fatalf("anchor on reed 5: %v", err)
	}
	c := s.Active().Curve
	if c == nil || c.ReedCount != 5 {
		t.Fatalf("curve = %+v, want a 5 reed curve", c)
	}
	if got := c.At(60)[4]; got != 12 {
		t.Fatalf("curve at reed 5 = %v, want 12", got)
	}
}

// Every rejection is an i18n-keyed ServiceError, the shape the frontend understands.
func TestErrorsAreKeyed(t *testing.T) {
	s := service(t)

	var se *ServiceError
	if err := s.SetA4(440); !errors.As(err, &se) || se.Key != "session.error.noSession" {
		t.Fatalf("SetA4 with nothing open: err = %v, want a keyed ErrNoSession", err)
	}
	if err := s.UpsertTake(take(69, 440, 0)); !errors.Is(err, ErrNoSession) {
		t.Fatalf("a reading with nothing open: err = %v, want %s", err, ErrNoSession.Key)
	}
	if _, err := s.Open("../etc/passwd"); !errors.As(err, &se) {
		t.Fatalf("a hostile id must come back keyed, got %v", err)
	}
	if _, err := s.Open(coresession.NewID()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("an unknown id: err = %v, want %s", err, ErrNotFound.Key)
	}
}
