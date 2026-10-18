package session

import (
	"errors"
	"testing"

	coresession "smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// The view is the goal line as the engine reads it and every treble reading on the curve's columns.
func TestCurveViewOverlaysTheReadings(t *testing.T) {
	s := service(t)
	if _, err := s.CurveView(); !errors.Is(err, ErrNoSession) {
		t.Fatalf("no session: err = %v, want %s", err, ErrNoSession.Key)
	}
	inst := coresession.Instrument{ReedCount: 3, A4: 440,
		Banks:     []coresession.Bank{coresession.BankM1, coresession.BankM2, coresession.BankM3},
		Registers: []coresession.Register{{Name: "M1M3", Banks: []coresession.Bank{coresession.BankM1, coresession.BankM3}}},
	}
	if _, err := s.Create(NewSessionDTO{Name: "Morino", Instrument: inst}); err != nil {
		t.Fatal(err)
	}
	view, err := s.CurveView()
	if err != nil || len(view.Goals) != 0 || len(view.Recorded) != 0 {
		t.Fatalf("view = %+v (%v), want empty without a curve or readings", view, err)
	}

	for col, v := range []float64{-8, 0, 8} {
		if err := s.SetAnchor(60, col, v, "reference"); err != nil {
			t.Fatal(err)
		}
		if err := s.SetAnchor(62, col, v*7, "reference"); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.SetRegister("M1M3"); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertTake(take(61, 440, -20, 30)); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertTake(take(70, 440, -1, 1)); err != nil {
		t.Fatal(err)
	}

	view, err = s.CurveView()
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Goals) != 3 || view.Goals[0].Note != 60 || view.Goals[2].Note != 62 {
		t.Fatalf("goal rows %+v, want the three notes the curve covers", view.Goals)
	}
	if g := view.Goals[2].Reeds; g[0] == nil || *g[0] != -56 || g[2] != nil {
		t.Fatalf("note 62 = %v, want -56 and no goal above the 50 unit limit", g)
	}
	if len(view.Recorded) != 2 {
		t.Fatalf("recorded = %+v, want two readings", view.Recorded)
	}
	in := view.Recorded[0]
	if in.Note != 61 || in.Register != "M1M3" || len(in.Reeds) != 2 {
		t.Fatalf("first reading = %+v", in)
	}
	if r := in.Reeds[1]; r.Column != 2 || r.Dev != 30 || r.Goal == nil || *r.Goal != 32 {
		t.Fatalf("M3 slot = %+v, want column 3 at 30 against a goal of 32", r)
	}
	if out := view.Recorded[1].Reeds[0]; out.Column != 0 || out.Goal != nil {
		t.Fatalf("a reading past the curve = %+v, want column 1 with no goal", out)
	}
}

// Fitting reads every slot in its card column, so a short register does not shift the ranks.
func TestFitCurveUsesTheCardColumns(t *testing.T) {
	s := service(t)
	inst := coresession.Instrument{ReedCount: 3, A4: 440,
		Banks:     []coresession.Bank{coresession.BankM1, coresession.BankM2, coresession.BankM3},
		Registers: []coresession.Register{{Name: "M1M3", Banks: []coresession.Bank{coresession.BankM1, coresession.BankM3}}},
	}
	if _, err := s.Create(NewSessionDTO{Name: "Morino", Instrument: inst}); err != nil {
		t.Fatal(err)
	}
	if err := s.SetRegister("M1M3"); err != nil {
		t.Fatal(err)
	}
	for n := tuning.Note(60); n <= 72; n++ {
		if err := s.UpsertTake(take(n, 440, -6, 6)); err != nil {
			t.Fatal(err)
		}
	}
	fit, err := s.FitCurve()
	if err != nil {
		t.Fatal(err)
	}
	at := fit.Curve.At(66)
	if at[0] < -6.5 || at[0] > -5.5 || at[1] != 0 || at[2] < 5.5 || at[2] > 6.5 {
		t.Fatalf("fitted goal at 66 = %v, want M1 -6, M2 untouched, M3 +6", at)
	}
}

// Loading a beat list from a session that has none, or an empty one, is refused.
func TestImportRefusesAnEmptyCurve(t *testing.T) {
	s := service(t)
	src := create(t, s, "Cleared", 3, 440)
	if err := s.SetAnchor(60, 0, 1, "reference"); err != nil {
		t.Fatal(err)
	}
	if err := s.ClearAnchor(60); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	create(t, s, "Target", 3, 440)
	if err := s.ImportCurve(src.ID); !errors.Is(err, ErrNoCurve) {
		t.Fatalf("import of an emptied curve: err = %v, want %s", err, ErrNoCurve.Key)
	}
}
