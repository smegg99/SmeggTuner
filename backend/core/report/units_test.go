package report

import (
	"errors"
	"strings"
	"testing"

	"smegg.me/smeggtuner/backend/core/target"
)

// Cents is a display choice: the numbers convert, the verdicts stay the backend's.
func TestTheSheetPrintsCentsWhenAsked(t *testing.T) {
	s := musette(t)
	ref := sheet(t, s)
	cents, err := Build(s, Options{Now: at, Unit: target.UnitCents})
	if err != nil {
		t.Fatal(err)
	}
	if ref.Unit != target.UnitReference || cents.Unit != target.UnitCents {
		t.Fatalf("units %q and %q", ref.Unit, cents.Unit)
	}

	r, c := row(t, ref, 60).Reeds[2], row(t, cents, 60).Reeds[2]
	if want := target.UnitCents.FromReference(r.Curr, 0); !near(c.Curr, want, 1e-9) {
		t.Errorf("Curr in cents = %v, want %v", c.Curr, want)
	}
	if !near(c.Error, c.Curr-c.Goal, 1e-9) || c.InTol != r.InTol {
		t.Errorf("the cents row %+v does not add up or changed its verdict from %+v", c, r)
	}
	if !near(cents.Summary.Tolerance, target.UnitCents.FromReference(target.DefaultTolerance, 0), 1e-9) {
		t.Errorf("tolerance printed as %v cents", cents.Summary.Tolerance)
	}

	if out := text(t, cents); !strings.Contains(out, "Tuning table (cents)") || !strings.Contains(out, "Cents from the tempered scale") {
		t.Error("a cents sheet does not say it is in cents")
	}
	if out := text(t, ref); !strings.Contains(out, "Tuning table (units of 0.06 %)") {
		t.Error("a reference sheet does not say what its unit is")
	}
}

func TestTheSheetRefusesAUnitItCannotPrint(t *testing.T) {
	if _, err := Build(musette(t), Options{Unit: target.UnitHz}); !errors.Is(err, ErrUnit) {
		t.Fatalf("Build in hz = %v, want ErrUnit", err)
	}
}
