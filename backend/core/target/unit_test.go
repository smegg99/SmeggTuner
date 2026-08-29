package target

import (
	"math"
	"testing"
)

// One reference unit is 0.06 percent of the note: about 1.04 cents, and the trip back is exact.
func TestUnitsAndCentsConvert(t *testing.T) {
	almost(t, UnitCents.FromReference(1, 0), 1200*math.Log2(1.0006), 1e-12, "one unit in cents")
	for _, d := range []float64{-50, -1, 0, 0.5, 1, 50} {
		almost(t, UnitCents.ToReference(UnitCents.FromReference(d, 0), 0), d, 1e-9, "units to cents and back")
		almost(t, UnitHz.ToReference(UnitHz.FromReference(d, 261.6), 261.6), d, 1e-9, "units to Hz and back")
		if UnitReference.FromReference(d, 440) != d || UnitReference.ToReference(d, 440) != d {
			t.Fatalf("reference units are the storage unit: %v changed", d)
		}
	}
	almost(t, UnitHz.FromReference(10, 440), 440*0.006, 1e-9, "ten units of A4 in Hz")
}

// A row restated in cents keeps its verdict and its Hz, and its error still adds up.
func TestAnErrorIsRestatedInTheDisplayUnit(t *testing.T) {
	e := Errors(measure(69, 440, 3), musette(t), 440, 0, nil)[0]
	c := e.In(UnitCents)
	almost(t, c.Curr, UnitCents.FromReference(e.Curr, 0), 1e-12, "curr in cents")
	almost(t, c.Error, c.Curr-c.Goal, 1e-12, "the error is curr minus goal in the shown unit")
	if c.InTol != e.InTol || c.CurrHz != e.CurrHz {
		t.Fatalf("the restatement changed the verdict or the Hz: %+v from %+v", c, e)
	}
	if h := e.In(UnitHz); h.Curr != e.CurrHz || h.Error != e.ErrorHz {
		t.Fatalf("an Hz restatement must read the Hz fields: %+v", h)
	}
	if e.In(UnitReference) != e {
		t.Fatal("reference units are what the row already holds")
	}

	b := BeatErrors(measure(72, 440, -6, 0, 6), musette(t), 440, 0, nil)[0]
	bc := b.In(UnitCents)
	almost(t, bc.Error, bc.Curr-bc.Goal, 1e-12, "a beat error adds up in cents too")
}
