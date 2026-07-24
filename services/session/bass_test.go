package session

import (
	"errors"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/dsp"
	coresession "smegg.me/smeggtuner/core/session"
)

func withBass() coresession.Instrument {
	i := musette()
	i.BassReeds = 5
	i.BassRegisters = []coresession.BassRegister{{Name: "Soft", Feet: []int{16, 8}}}
	return i
}

func openWithBass(t *testing.T) *Service {
	t.Helper()
	s := service(t)
	if _, err := s.Create(NewSessionDTO{Name: "Bass", Instrument: withBass()}); err != nil {
		t.Fatal(err)
	}
	return s
}

// Turning the bench toward the bass side swaps the goal onto the machine's whole ladder; a pulled
// bass switch narrows it; back to treble restores the register's banks.
func TestBassBenchImposesTheLadder(t *testing.T) {
	s := openWithBass(t)

	if err := s.SetBass(true); err != nil {
		t.Fatal(err)
	}
	b := bench(t, s)
	if !b.Bass || b.Reeds != 5 || len(b.BassFeet) != 5 || b.BassFeet[0] != 32 {
		t.Fatalf("bass bench = %+v, want the whole 32..2 ladder", b)
	}
	if g := s.Goal(); len(g.BassFeet) != 5 || g.Banks != nil {
		t.Fatalf("goal = feet %v banks %v, want the ladder and no banks", g.BassFeet, g.Banks)
	}

	if err := s.SetBassRegister("Soft"); err != nil {
		t.Fatal(err)
	}
	if b := bench(t, s); len(b.BassFeet) != 2 || b.BassFeet[0] != 16 || b.Reeds != 2 {
		t.Fatalf("bass bench = %+v, want the switch's 16.8", b)
	}
	if err := s.SetBassRegister("Loud"); !errors.Is(err, ErrNoBassRegister) {
		t.Fatalf("an unknown bass switch must be refused, got %v", err)
	}

	if err := s.SetBass(false); err != nil {
		t.Fatal(err)
	}
	if b := bench(t, s); b.Bass || len(b.Banks) != 3 {
		t.Fatalf("treble bench = %+v, want the register's banks back", b)
	}
}

func TestBassNeedsADeclaredMachine(t *testing.T) {
	s := openMusette(t)
	if err := s.SetBass(true); !errors.Is(err, ErrNoBassMachine) {
		t.Fatalf("a bench without a bass machine must refuse the turn, got %v", err)
	}
}

// A take recorded while the bench faces the bass side is stamped as one, under the pulled switch,
// and never collides with a treble take of the same note.
func TestBassTakesAreStampedAsBass(t *testing.T) {
	s := openWithBass(t)
	if err := s.UpsertTake(coresession.Take{Note: 45, At: time.Unix(1, 0),
		Reeds: []dsp.ReedMeasure{{Freq: 110}}}); err != nil {
		t.Fatal(err)
	}
	if err := s.SetBass(true); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertTake(coresession.Take{Note: 45, At: time.Unix(2, 0),
		Reeds: []dsp.ReedMeasure{{Freq: 110}}}); err != nil {
		t.Fatal(err)
	}

	d, err := s.Data()
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Takes) != 2 {
		t.Fatalf("a bass and a treble take of one note are different voices, got %d", len(d.Takes))
	}
	var bass *coresession.Take
	for i := range d.Takes {
		if d.Takes[i].Bass {
			bass = &d.Takes[i]
		}
	}
	if bass == nil {
		t.Fatal("no take was stamped as bass")
	}
}
