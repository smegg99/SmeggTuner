package repositories_test

import (
	"errors"
	"testing"

	"smegg.me/smeggtuner/core/datastore/datastoretest"
	"smegg.me/smeggtuner/core/repositories"
	"smegg.me/smeggtuner/core/session"
)

// A fresh install ships no bundled instruments.
func TestAFreshShelfIsEmpty(t *testing.T) {
	datastoretest.Init(t)

	all, err := repositories.GetInstrumentRepository().List()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 0 {
		t.Fatalf("a fresh install ships %d instruments, want none: %+v", len(all), all)
	}
}

func TestAnInstrumentGoesOnTheShelf(t *testing.T) {
	datastoretest.Init(t)
	repo := repositories.GetInstrumentRepository()

	mine := &session.Template{
		Name:       "Zupan Alpe IV",
		Instrument: session.Instrument{ReedCount: 4, Banks: []session.Bank{session.BankL, session.BankM1, session.BankM2, session.BankM3}},
	}
	if err := repo.Save(mine); err != nil {
		t.Fatal(err)
	}
	if mine.ID == "" {
		t.Fatal("a saved instrument has no id")
	}

	all, err := repo.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 || all[0].ID != mine.ID {
		t.Fatalf("the shelf holds %+v", all)
	}
	if all[0].HasImage {
		t.Fatal("an instrument nobody photographed claims to have a photograph")
	}
}

func TestAnInstrumentNeedsAName(t *testing.T) {
	datastoretest.Init(t)
	repo := repositories.GetInstrumentRepository()

	if err := repo.Save(&session.Template{Name: "  ", Instrument: session.Instrument{ReedCount: 3}}); !errors.Is(err, session.ErrTemplateName) {
		t.Fatalf("save = %v, want %v", err, session.ErrTemplateName)
	}
}

func TestAnInstrumentKeepsItsReferencePitch(t *testing.T) {
	datastoretest.Init(t)
	repo := repositories.GetInstrumentRepository()

	mine := &session.Template{
		Name:       "Hohner, old",
		Instrument: session.Instrument{ReedCount: 3, Banks: []session.Bank{session.BankM1, session.BankM2, session.BankM3}, A4: 442},
	}
	if err := repo.Save(mine); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Get(mine.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Instrument.A4 != 442 {
		t.Fatalf("A4 = %v, want 442", got.Instrument.A4)
	}
}
