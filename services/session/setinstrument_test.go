package session

import (
	"errors"
	"testing"

	coresession "smegg.me/smeggtuner/core/session"
)

// The accordion is described where the work happens, so the description has to be
// able to change while the work is going on.
func TestTheAccordionIsDescribedWhileItIsOnTheBench(t *testing.T) {
	s := openMusette(t)

	next := musette()
	next.Name = "Bugari 508/ARS"
	next.Lo, next.Hi = 41, 100
	next.Tolerance = 2.5

	if err := s.SetInstrument(next); err != nil {
		t.Fatalf("SetInstrument: %v", err)
	}

	got := s.Active()
	if got.Instrument.Name != "Bugari 508/ARS" {
		t.Errorf("name = %q, want the one just given", got.Instrument.Name)
	}
	if got.Instrument.Tolerance != 2.5 {
		t.Errorf("tolerance = %v, want 2.5", got.Instrument.Tolerance)
	}
}

// A switch the technician had not described yet is the commonest mid-session
// correction there is, and it moves no column.
func TestASwitchMayBeDescribedAfterTheReadingsStart(t *testing.T) {
	s := openMusette(t)
	dto := s.Active()
	if err := s.UpsertTake(take(69, dto.A4, 0)); err != nil {
		t.Fatalf("UpsertTake: %v", err)
	}

	next := musette()
	next.Registers = append(next.Registers, coresession.Register{
		Name: "M3", Banks: []coresession.Bank{coresession.BankM3},
	})

	if err := s.SetInstrument(next); err != nil {
		t.Fatalf("adding a switch with readings present: %v", err)
	}
	if n := len(s.Active().Instrument.Registers); n != 4 {
		t.Errorf("registers = %d, want 4", n)
	}
}

// The card's columns are what every reading was filed under. Moving them would
// silently re-file readings nobody re-measured.
func TestTheCardsColumnsFreezeOnceThereAreReadings(t *testing.T) {
	cases := map[string]func(coresession.Instrument) coresession.Instrument{
		"a rank added": func(i coresession.Instrument) coresession.Instrument {
			i.Banks = append(append([]coresession.Bank(nil), i.Banks...), coresession.BankH)
			i.Registers = append(i.Registers, coresession.Register{
				Name: "MMMH",
				Banks: []coresession.Bank{
					coresession.BankM1, coresession.BankM2, coresession.BankM3, coresession.BankH,
				},
			})
			i.ReedCount = 4
			return i
		},
		"a rank dropped": func(i coresession.Instrument) coresession.Instrument {
			i.Banks = []coresession.Bank{coresession.BankM1, coresession.BankM2}
			i.Registers = []coresession.Register{{
				Name:  "M1M2",
				Banks: []coresession.Bank{coresession.BankM1, coresession.BankM2},
			}}
			i.ReedCount = 2
			return i
		},
		"a bass machine declared": func(i coresession.Instrument) coresession.Instrument {
			i.BassReeds = 4
			return i
		},
	}

	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			s := openMusette(t)
			dto := s.Active()
			if err := s.UpsertTake(take(69, dto.A4, 0)); err != nil {
				t.Fatalf("UpsertTake: %v", err)
			}

			if err := s.SetInstrument(change(musette())); !errors.Is(err, ErrHasReadings) {
				t.Errorf("%s with readings = %v, want ErrHasReadings", name, err)
			}
		})
	}
}

func TestTheColumnsAreFreeUntilTheFirstReading(t *testing.T) {
	s := openMusette(t)

	next := musette()
	next.Banks = append(next.Banks, coresession.BankH)
	next.Registers = append(next.Registers, coresession.Register{
		Name: "MMMH",
		Banks: []coresession.Bank{
			coresession.BankM1, coresession.BankM2, coresession.BankM3, coresession.BankH,
		},
	})
	next.ReedCount = 4

	if err := s.SetInstrument(next); err != nil {
		t.Fatalf("widening an empty session: %v", err)
	}
	if n := s.Active().Instrument.ReedCount; n != 4 {
		t.Errorf("reed count = %d, want 4", n)
	}
}

func TestDescribingTheAccordionNeedsAnOpenSession(t *testing.T) {
	s := service(t)
	if err := s.SetInstrument(musette()); !errors.Is(err, ErrNoSession) {
		t.Errorf("SetInstrument with no session = %v, want ErrNoSession", err)
	}
}

func TestAnAccordionThatDoesNotAddUpIsRefused(t *testing.T) {
	s := openMusette(t)

	bad := musette()
	// A switch sounding a rank the instrument does not have would put a reading in
	// a column the card has not got.
	bad.Registers = append(bad.Registers, coresession.Register{
		Name: "H", Banks: []coresession.Bank{coresession.BankH},
	})

	if err := s.SetInstrument(bad); err == nil {
		t.Error("a switch naming a rank the instrument lacks was accepted")
	}
}

// Nothing handed out may share memory with the active session.
func TestTheDescriptionIsCopiedIn(t *testing.T) {
	s := openMusette(t)

	next := musette()
	if err := s.SetInstrument(next); err != nil {
		t.Fatalf("SetInstrument: %v", err)
	}
	next.Banks[0] = coresession.BankH
	next.Registers[0].Banks[0] = coresession.BankH

	got := s.Active().Instrument
	if got.Banks[0] != coresession.BankM1 || got.Registers[0].Banks[0] != coresession.BankM1 {
		t.Error("editing the caller's instrument reached into the open session")
	}
}

// Remembering follows a description that has only just been typed. The store is
// written asynchronously (store.go, touch/saver), so reading the accordion back off
// disk would remember whatever was there before the last keystroke.
func TestRememberingTakesTheAccordionAsItStandsNow(t *testing.T) {
	s := openMusette(t)
	dto := s.Active()

	next := musette()
	next.Name = "Bugari 508/ARS"
	next.Tolerance = 2.5
	if err := s.SetInstrument(next); err != nil {
		t.Fatalf("SetInstrument: %v", err)
	}

	got, err := s.AdoptInstrument(dto.ID, "Bugari 508/ARS")
	if err != nil {
		t.Fatalf("adopt: %v", err)
	}
	if got.Instrument.Tolerance != 2.5 {
		t.Errorf("tolerance = %v, want the 2.5 just set; it read a stale copy off disk",
			got.Instrument.Tolerance)
	}
}

// The link has to land on the open session too, not only on the copy written out:
// the next flush writes the in-memory one over it.
func TestRememberingLinksTheOpenSessionAndSurvivesTheNextWrite(t *testing.T) {
	s := openMusette(t)
	dto := s.Active()

	got, err := s.AdoptInstrument(dto.ID, "Bugari 508/ARS")
	if err != nil {
		t.Fatalf("adopt: %v", err)
	}
	if id := s.Active().InstrumentID; id != got.ID {
		t.Fatalf("the open session's link = %q, want %q", id, got.ID)
	}

	if err := s.Save(); err != nil {
		t.Fatalf("flush: %v", err)
	}
	back, err := s.Open(dto.ID)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if back.InstrumentID != got.ID {
		t.Errorf("after a flush the link = %q, want %q", back.InstrumentID, got.ID)
	}
}
