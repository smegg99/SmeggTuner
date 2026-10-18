package session

import (
	"errors"
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/target"
	"smegg.me/smeggtuner/backend/core/tuning"
)

func TestCurveAnchorsClearAndDrop(t *testing.T) {
	s := service(t)
	create(t, s, "Morino", 3, 440)

	if err := s.SetAnchor(60, 2, 8, "reference"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetAnchor(72, 2, 16, "reference"); err != nil {
		t.Fatal(err)
	}
	c := s.Active().Curve
	if len(c.Anchors) != 2 {
		t.Fatalf("anchors = %d, want 2", len(c.Anchors))
	}
	if got := c.At(66)[2]; got != 12 {
		t.Fatalf("interpolated goal = %v, want 12 halfway between the anchors", got)
	}

	// Hz and cents are stored as reference units, Hz read against A4 = 440 Hz, so the display unit can change after.
	if err := s.SetAnchor(69, 1, 2, "hz"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetAnchor(69, 0, 2, "cent"); err != nil {
		t.Fatal(err)
	}
	at := s.Active().Curve.At(69)
	if want := target.UnitHz.ToGoal(2, 69); at[1] != want {
		t.Fatalf("hz anchor stored as %v units, want %v", at[1], want)
	}
	if want := target.UnitCents.ToReference(2, 0); at[0] != want {
		t.Fatalf("cent anchor stored as %v units, want %v", at[0], want)
	}
	if err := s.SetAnchor(69, 1, 2, "furlong"); !errors.Is(err, ErrInvalidUnit) {
		t.Fatalf("unknown unit: err = %v, want %s", err, ErrInvalidUnit.Key)
	}

	if err := s.ClearAnchor(69); err != nil {
		t.Fatal(err)
	}
	if len(s.Active().Curve.Anchors) != 2 {
		t.Fatal("clearing a note must drop exactly its anchor")
	}

	if err := s.DropCurve(); err != nil {
		t.Fatal(err)
	}
	if s.Active().Curve != nil {
		t.Fatal("a dropped curve leaves the session with no goal, which is a state")
	}
	if g := s.Goal(); g.Curve != nil || g.A4 != 440 {
		t.Fatalf("Goal() = %+v, want no curve and the session's reference", g)
	}
}

func TestFitCurveFromAPass(t *testing.T) {
	s := service(t)
	create(t, s, "Morino", 3, 440)

	// A tremolo the instrument holds: reed 3 eight units sharp everywhere, reed 1 eight flat.
	for n := tuning.Note(48); n <= 84; n++ {
		if err := s.UpsertTake(take(n, 440, -8, 0, 8)); err != nil {
			t.Fatal(err)
		}
	}

	fit, err := s.FitCurve()
	if err != nil {
		t.Fatal(err)
	}
	if fit.Used == 0 || fit.Curve == nil {
		t.Fatalf("fit = %+v, want a curve fitted to the pass", fit)
	}
	goal := s.Goal().Curve
	if goal == nil {
		t.Fatal("a fitted curve becomes the session's goal")
	}
	for _, n := range []tuning.Note{50, 60, 70, 80} {
		at := goal.At(n)
		if diff := at[2] - 8; diff > 1 || diff < -1 {
			t.Fatalf("fitted goal for reed 3 at note %d = %v, want the instrument's own +8", n, at[2])
		}
	}

	// Fitting a session with nothing in it is a refusal, not a curve of zeros.
	if err := s.ClearTakes(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.FitCurve(); !errors.Is(err, ErrNoReadings) {
		t.Fatalf("fit of an empty session: err = %v, want %s", err, ErrNoReadings.Key)
	}
}

func TestImportCurveFromAnotherSession(t *testing.T) {
	s := service(t)
	src := create(t, s, "Source", 3, 440)
	if err := s.SetAnchor(60, 2, 10, "reference"); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}

	create(t, s, "Target", 3, 440)
	if s.Active().Curve != nil {
		t.Fatal("a new session starts with no goal")
	}
	if err := s.ImportCurve(src.ID); err != nil {
		t.Fatal(err)
	}
	if got := s.Active().Curve.At(60)[2]; got != 10 {
		t.Fatalf("imported goal = %v, want the source session's 10", got)
	}

	empty := create(t, s, "Empty", 3, 440)
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	if err := s.ImportCurve(empty.ID); !errors.Is(err, ErrNoCurve) {
		t.Fatalf("import from a session with no goal: err = %v, want %s", err, ErrNoCurve.Key)
	}
}

// A new curve interpolates and does not extrapolate, as the reference beat list; the flags persist.
func TestCurveFlagsDefaultAndPersist(t *testing.T) {
	s := service(t)
	dto := create(t, s, "Morino", 3, 440)

	if err := s.SetAnchor(60, 2, 4, "reference"); err != nil {
		t.Fatal(err)
	}
	if err := s.SetAnchor(72, 2, 16, "reference"); err != nil {
		t.Fatal(err)
	}
	c := s.Active().Curve
	if !c.Interpolate || c.ExtrapolateLeft || c.ExtrapolateRight {
		t.Fatalf("a new curve came up with %+v, want interpolation on and extrapolation off", c)
	}
	if got := c.At(66)[2]; got != 10 {
		t.Fatalf("goal halfway between the anchors = %v, want 10", got)
	}

	if err := s.SetInterpolate(false); err != nil {
		t.Fatal(err)
	}
	if err := s.SetExtrapolateLeft(true); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Open(dto.ID); err != nil {
		t.Fatal(err)
	}

	c = s.Active().Curve
	if c.Interpolate || c.ExtrapolateRight || !c.ExtrapolateLeft {
		t.Fatalf("the flags changed across a save: %+v", c)
	}
	if got := c.At(66)[2]; got != 4 {
		t.Fatalf("interpolation off: goal at 66 = %v, want the nearer anchor's 4", got)
	}
	if got := c.At(100)[2]; got != meter.Missing {
		t.Fatalf("extrapolation off: goal above the last anchor = %v, want none", got)
	}
	if got := c.At(56)[2]; got != 0 {
		t.Fatalf("extrapolation left: the line 4 -> 16 reaches zero at 56, goal = %v, want 0", got)
	}
	if got := c.At(58)[2]; got != 2 {
		t.Fatalf("extrapolation left: goal at 58 = %v, want the line's 2", got)
	}

	if err := s.SetInterpolate(true); err != nil {
		t.Fatal(err)
	}
	if got := s.Active().Curve.At(66)[2]; got != 10 {
		t.Fatalf("interpolation back on: goal at 66 = %v, want 10", got)
	}
}
