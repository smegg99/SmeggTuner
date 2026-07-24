package repositories_test

import (
	"errors"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/datastore/datastoretest"
	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/repositories"
	"smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/target"
	"smegg.me/smeggtuner/core/tuning"
)

// recorded is a fully populated session, so a round trip that drops a JSON column has something to lose.
func recorded(t *testing.T) *session.Session {
	t.Helper()
	s := session.New("Jan K.", session.Instrument{
		Banks:     []session.Bank{session.BankM1, session.BankM2, session.BankM3},
		Registers: []session.Register{{Name: "MMM", Banks: []session.Bank{session.BankM1, session.BankM2, session.BankM3}}},
		ReedCount: 3,
	}, 442)
	s.Curve = &target.Curve{
		ReedCount: 3,
		Unit:      target.UnitCents,
		Anchors:   []target.Anchor{{Note: tuning.Note(60), Reeds: []float64{0, -4, 4}}},
	}
	s.UpsertTake(session.Take{
		Note:  tuning.Note(60),
		At:    time.Now(),
		Reeds: []dsp.ReedMeasure{{Freq: 261.6, DevCents: 0.4}},
	})
	return s
}

func TestASessionSurvivesTheDatabase(t *testing.T) {
	datastoretest.Init(t)
	repo := repositories.GetSessionRepository()

	s := recorded(t)
	if err := repo.Save(s); err != nil {
		t.Fatal(err)
	}

	got, err := repo.Get(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != s.Name || got.A4 != s.A4 {
		t.Fatalf("got %q at %v, want %q at %v", got.Name, got.A4, s.Name, s.A4)
	}
	if len(got.Instrument.Registers) != 1 {
		t.Fatalf("the instrument did not survive: %+v", got.Instrument)
	}
	if got.Curve == nil || len(got.Curve.Anchors) != 1 {
		t.Fatalf("the goal did not survive: %+v", got.Curve)
	}
	if len(got.Takes) != 1 || got.Takes[0].Reeds[0].Freq != 261.6 {
		t.Fatalf("the readings did not survive: %+v", got.Takes)
	}
}

func TestTheListIsNewestFirst(t *testing.T) {
	datastoretest.Init(t)
	repo := repositories.GetSessionRepository()

	first := recorded(t)
	if err := repo.Save(first); err != nil {
		t.Fatal(err)
	}
	second := recorded(t)
	second.Name = "Marek W."
	if err := repo.Save(second); err != nil {
		t.Fatal(err)
	}

	all, err := repo.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 || all[0].ID != second.ID {
		t.Fatalf("list = %+v, want the newest first", all)
	}
	if all[0].Readings != 1 || !all[0].HasCurve {
		t.Fatalf("the summary miscounts the session: %+v", all[0])
	}
}

func TestSaveStampsUpdatedAndInsertDoesNot(t *testing.T) {
	datastoretest.Init(t)
	repo := repositories.GetSessionRepository()

	old := recorded(t)
	then := time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC)
	old.Created, old.Updated = then, then
	if err := repo.Insert(old); err != nil {
		t.Fatal(err)
	}
	got, err := repo.Get(old.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Updated.Equal(then) {
		t.Fatalf("Insert stamped Updated: %v", got.Updated)
	}

	if err := repo.Save(got); err != nil {
		t.Fatal(err)
	}
	if saved, _ := repo.Get(old.ID); !saved.Updated.After(then) {
		t.Fatalf("Save did not stamp Updated: %v", saved.Updated)
	}
}

func TestAMissingSessionIsNotFound(t *testing.T) {
	datastoretest.Init(t)
	repo := repositories.GetSessionRepository()

	if _, err := repo.Get(session.NewID()); !errors.Is(err, repositories.ErrNotFound) {
		t.Fatalf("get = %v, want %v", err, repositories.ErrNotFound)
	}
	if err := repo.Delete(session.NewID()); !errors.Is(err, repositories.ErrNotFound) {
		t.Fatalf("delete = %v, want %v", err, repositories.ErrNotFound)
	}
	// A malformed id from the frontend names nothing: same answer.
	if _, err := repo.Get("../../etc/passwd"); !errors.Is(err, repositories.ErrNotFound) {
		t.Fatalf("get = %v, want %v", err, repositories.ErrNotFound)
	}
}

func TestDeleteTakesOneSession(t *testing.T) {
	datastoretest.Init(t)
	repo := repositories.GetSessionRepository()

	keep, drop := recorded(t), recorded(t)
	if err := repo.Save(keep); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(drop); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(drop.ID); err != nil {
		t.Fatal(err)
	}

	all, err := repo.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 || all[0].ID != keep.ID {
		t.Fatalf("list after delete = %+v", all)
	}
}
