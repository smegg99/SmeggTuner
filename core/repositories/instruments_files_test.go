package repositories_test

import (
	"path/filepath"
	"testing"

	"smegg.me/smeggtuner/core/datastore/datastoretest"
	"smegg.me/smeggtuner/core/repositories"
	"smegg.me/smeggtuner/core/session"
)

// An imported instrument always gets a NEW id, so it can't overwrite an existing one.
func TestAnImportedInstrumentDoesNotOverwriteMine(t *testing.T) {
	datastoretest.Init(t)
	repo := repositories.GetInstrumentRepository()

	mine := &session.Template{Name: "Morino", Instrument: session.Instrument{ReedCount: 3, Banks: []session.Bank{session.BankM1, session.BankM2, session.BankM3}}}
	if err := repo.Save(mine); err != nil {
		t.Fatal(err)
	}

	// A file that happens to carry the same id as mine.
	theirs := &session.Template{ID: mine.ID, Name: "Their Morino", Instrument: session.Instrument{ReedCount: 2, Banks: []session.Bank{session.BankL, session.BankM1}}}
	path := filepath.Join(t.TempDir(), "theirs"+session.InstrumentFileExt)
	if err := session.WriteInstrumentFile(path, theirs, nil); err != nil {
		t.Fatal(err)
	}

	got, err := repo.ImportFile(path)
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if got.ID == mine.ID {
		t.Fatal("an import overwrote an instrument that was already there")
	}

	back, err := repo.Get(mine.ID)
	if err != nil {
		t.Fatalf("my own instrument is gone: %v", err)
	}
	if back.Name != "Morino" || back.Instrument.ReedCount != 3 {
		t.Fatalf("my own instrument was changed under me: %+v", back)
	}
}

func TestAPhotographSurvivesTheStifRoundTrip(t *testing.T) {
	datastoretest.Init(t)
	repo := repositories.GetInstrumentRepository()

	mine := &session.Template{Name: "Morino", Instrument: session.Instrument{ReedCount: 3}}
	if err := repo.Save(mine); err != nil {
		t.Fatal(err)
	}
	if err := repo.SetImage(mine.ID, photo(t, 500, 400)); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "travel"+session.InstrumentFileExt)
	if err := repo.ExportFile(mine.ID, path); err != nil {
		t.Fatal(err)
	}
	got, err := repo.ImportFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !got.HasImage {
		t.Fatal("the photograph did not survive the trip")
	}
	if jpg, _, err := repo.Image(got.ID); err != nil || len(jpg) == 0 {
		t.Fatalf("the imported photograph cannot be read back: %v", err)
	}
}
