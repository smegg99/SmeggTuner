package repositories_test

import (
	"errors"
	"testing"
	"time"

	"smegg.me/smeggtuner/backend/core/datastore/datastoretest"
	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/repositories"
	"smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/target"
	"smegg.me/smeggtuner/backend/core/tuning"
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
		Unit:      target.UnitReference,
		Anchors:   []target.Anchor{{Note: tuning.Note(60), Reeds: []float64{0, -4, 4}}},
	}
	s.UpsertTake(session.Take{
		Note:  tuning.Note(60),
		At:    time.Now(),
		Reeds: []dsp.ReedMeasure{{Present: true, Freq: 261.6, Dev: 0.4}, {Dev: -99}, {Present: true, Freq: 262.1, Dev: 3.5}},
		Beats: []dsp.BeatMeasure{{Pair: "1-2"}, {Pair: "2-3"}},
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
	if len(got.Takes) != 1 || got.Takes[0].Reeds[0].Freq != 261.6 || got.Takes[0].Reeds[0].Dev != 0.4 {
		t.Fatalf("the readings did not survive: %+v", got.Takes)
	}
	if r := got.Takes[0].Reeds; len(r) != 3 || r[1].Present || !r[2].Present || len(got.Takes[0].Beats) != 2 {
		t.Fatalf("the slots did not keep their places: %+v", got.Takes[0])
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

func TestSaveStampsUpdated(t *testing.T) {
	datastoretest.Init(t)
	repo := repositories.GetSessionRepository()

	s := recorded(t)
	then := time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC)
	s.Created, s.Updated = then, then
	if err := repo.Save(s); err != nil {
		t.Fatal(err)
	}
	saved, err := repo.Get(s.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !saved.Updated.After(then) {
		t.Fatalf("Save did not stamp Updated: %v", saved.Updated)
	}
	if !saved.Created.Equal(then) {
		t.Fatalf("Save touched Created: %v", saved.Created)
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
