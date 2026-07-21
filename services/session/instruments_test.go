package session

import (
	"errors"
	"path/filepath"
	"slices"
	"testing"

	coresession "smegg.me/smeggtuner/core/session"
)

// The library must live in the service's datastore, not the real config directory.
func TestTheInstrumentLibrarySitsBesideTheSessions(t *testing.T) {
	s := service(t)

	create(t, s, "Bench", 3, 442)
	if _, err := s.SaveInstrument("Mine"); err != nil {
		t.Fatal(err)
	}

	all, err := s.templates().List()
	if err != nil {
		t.Fatalf("the instrument was not saved beside the sessions: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("the library holds %d instruments", len(all))
	}
}

// The library starts empty; no default instrument to be picked by accident.
func TestAFreshBenchHasNoInstruments(t *testing.T) {
	s := service(t)
	all, err := s.Instruments()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 0 {
		t.Fatalf("a fresh install offers %d instruments, want none: %+v", len(all), all)
	}
}

// Saving keeps the model - banks, registers - not the serial.
func TestSavingTheInstrumentOnTheBench(t *testing.T) {
	s := service(t)
	if _, err := s.Create(NewSessionDTO{
		Name: "Jan K.",
		Instrument: coresession.Instrument{
			Make:      "Hohner",
			Model:     "Morino",
			Serial:    "12345",
			Banks:     []coresession.Bank{coresession.BankM1, coresession.BankM2, coresession.BankM3},
			Registers: []coresession.Register{{Name: "MMM", Banks: []coresession.Bank{coresession.BankM1, coresession.BankM2, coresession.BankM3}}},
			ReedCount: 3,
		},
	}); err != nil {
		t.Fatal(err)
	}

	got, err := s.SaveInstrument("Morino, as I set it up")
	if err != nil {
		t.Fatal(err)
	}
	if got.Instrument.Serial != "" {
		t.Fatal("the template kept the serial of one particular accordion")
	}
	if !slices.Equal(got.Instrument.Banks, []coresession.Bank{coresession.BankM1, coresession.BankM2, coresession.BankM3}) {
		t.Fatalf("banks = %v", got.Instrument.Banks)
	}

	all, err := s.Instruments()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 || all[0].ID != got.ID {
		t.Fatalf("it is not in the library: %+v", all)
	}
}

func TestSavingAnInstrumentNeedsAnInstrument(t *testing.T) {
	s := service(t)
	if _, err := s.SaveInstrument("Nothing"); !errors.Is(err, ErrNoSession) {
		t.Fatalf("save = %v, want %v", err, ErrNoSession)
	}
}

// An imported session's instrument can be adopted afterwards; it carries no photograph.
func TestAnInstrumentCanBeTakenOffAnImportedSession(t *testing.T) {
	s := service(t)
	dto := create(t, s, "Jan K.", 3, 442)

	got, err := s.AdoptInstrument(dto.ID, "Their Morino")
	if err != nil {
		t.Fatalf("adopt: %v", err)
	}
	if got.Instrument.ReedCount != 3 {
		t.Fatalf("the instrument did not come across: %+v", got.Instrument)
	}
	if got.HasImage {
		t.Fatal("an adopted instrument invented a photograph")
	}

	all, err := s.Instruments()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 || all[0].ID != got.ID {
		t.Fatalf("it is not on the shelf: %+v", all)
	}

	// The session now knows its instrument, so it never re-offers adoption.
	back, err := s.Open(dto.ID)
	if err != nil {
		t.Fatal(err)
	}
	if back.InstrumentID != got.ID {
		t.Fatalf("the session still does not know its instrument: %q", back.InstrumentID)
	}
}

func TestAnInstrumentTravelsBetweenBenches(t *testing.T) {
	s := service(t)
	create(t, s, "Bench", 3, 442)
	saved, err := s.SaveInstrument("Castagnari")
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "out.stif")
	if err := s.ExportInstrument(saved.ID, path); err != nil {
		t.Fatalf("export: %v", err)
	}

	got, err := s.ImportInstrument(path)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if got.Name != "Castagnari" {
		t.Fatalf("name = %q", got.Name)
	}
	if got.ID == saved.ID {
		t.Fatal("the import landed on top of the instrument it came from")
	}
}
