package session

import (
	"errors"
	"testing"

	coresession "smegg.me/smeggtuner/core/session"
)

func TestLearningTheKeyboardRangeKeepsItOnTheShelf(t *testing.T) {
	s := service(t)

	tpl, err := s.SaveInstrumentSpec(coresession.Template{
		Name:       "Morino",
		Instrument: coresession.Instrument{ReedCount: 3},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create(NewSessionDTO{
		Name: "Jan K.", InstrumentID: tpl.ID,
		Instrument: coresession.Instrument{ReedCount: 3},
	}); err != nil {
		t.Fatal(err)
	}

	if err := s.SetKeyboardRange(36, 84); err != nil {
		t.Fatalf("set range: %v", err)
	}

	if a := s.Active(); a.Instrument.Lo != 36 || a.Instrument.Hi != 84 {
		t.Fatalf("session range = %d..%d", a.Instrument.Lo, a.Instrument.Hi)
	}

	back, err := s.templates().Get(tpl.ID)
	if err != nil {
		t.Fatal(err)
	}
	if back.Instrument.Lo != 36 || back.Instrument.Hi != 84 {
		t.Fatalf("shelf range = %d..%d, want 36..84", back.Instrument.Lo, back.Instrument.Hi)
	}
}

func TestABackwardsKeyboardRangeIsRefused(t *testing.T) {
	s := service(t)
	create(t, s, "Bench", 3, 442)
	if err := s.SetKeyboardRange(84, 36); !errors.Is(err, ErrInvalidInstrument) {
		t.Fatalf("set = %v, want %v", err, ErrInvalidInstrument)
	}
}
