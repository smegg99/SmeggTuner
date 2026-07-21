package repositories_test

import (
	"errors"
	"testing"

	"smegg.me/smeggtuner/core/datastore/datastoretest"
	"smegg.me/smeggtuner/core/repositories"
	"smegg.me/smeggtuner/core/session"
)

func TestAnInstrumentKeepsItsPhotograph(t *testing.T) {
	datastoretest.Init(t)
	repo := repositories.GetInstrumentRepository()

	mine := &session.Template{Name: "Morino", Instrument: session.Instrument{ReedCount: 3}}
	if err := repo.Save(mine); err != nil {
		t.Fatal(err)
	}
	if err := repo.SetImage(mine.ID, photo(t, 600, 400)); err != nil {
		t.Fatal(err)
	}

	all, err := repo.List()
	if err != nil {
		t.Fatal(err)
	}
	if !all[0].HasImage || all[0].ImageRev == 0 {
		t.Fatalf("the shelf does not know the instrument has a photograph: %+v", all[0])
	}
	jpg, rev, err := repo.Image(mine.ID)
	if err != nil || len(jpg) == 0 || rev != all[0].ImageRev {
		t.Fatalf("the photograph cannot be read back: %v (%d bytes, rev %d)", err, len(jpg), rev)
	}

	if err := repo.SetImage(mine.ID, nil); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.Image(mine.ID); !errors.Is(err, session.ErrNoImage) {
		t.Fatalf("the photograph is still there: %v", err)
	}
}

// A description edit must not touch the image.
func TestEditingTheDescriptionKeepsThePhotograph(t *testing.T) {
	datastoretest.Init(t)
	repo := repositories.GetInstrumentRepository()

	mine := &session.Template{Name: "Morino", Instrument: session.Instrument{ReedCount: 3}}
	if err := repo.Save(mine); err != nil {
		t.Fatal(err)
	}
	if err := repo.SetImage(mine.ID, photo(t, 300, 300)); err != nil {
		t.Fatal(err)
	}

	edit := &session.Template{ID: mine.ID, Name: "Morino VI M", Instrument: session.Instrument{ReedCount: 3}}
	if err := repo.Save(edit); err != nil {
		t.Fatal(err)
	}
	if !edit.HasImage {
		t.Fatal("Save did not hand back the photograph facts from the row")
	}
	if _, _, err := repo.Image(mine.ID); err != nil {
		t.Fatalf("editing the description lost the photograph: %v", err)
	}
}

func TestDeletingAnInstrumentTakesItsPhotographToo(t *testing.T) {
	datastoretest.Init(t)
	repo := repositories.GetInstrumentRepository()

	mine := &session.Template{Name: "Morino", Instrument: session.Instrument{ReedCount: 3}}
	if err := repo.Save(mine); err != nil {
		t.Fatal(err)
	}
	if err := repo.SetImage(mine.ID, photo(t, 300, 300)); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(mine.ID); err != nil {
		t.Fatal(err)
	}

	if _, err := repo.Get(mine.ID); !errors.Is(err, session.ErrNoTemplate) {
		t.Fatalf("get after delete = %v, want %v", err, session.ErrNoTemplate)
	}
	if _, _, err := repo.Image(mine.ID); !errors.Is(err, session.ErrNoTemplate) {
		t.Fatalf("image after delete = %v, want %v", err, session.ErrNoTemplate)
	}
}
