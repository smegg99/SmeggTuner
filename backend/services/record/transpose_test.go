package record

import (
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// The table states the transposition and reads Hz at the sounding pitch; the row stays the key.
func TestATransposedPassIsReadAtItsSoundingPitch(t *testing.T) {
	s, sessions := services(t)
	open(t, sessions, 1)
	if err := sessions.SetTranspose(3); err != nil {
		t.Fatal(err)
	}
	s.SetArmed(true)

	sounding := tuning.Note(72).Freq(a4) // the A4 key of an instrument whose C sounds as D sharp
	s.OnMeasurement(measured(69, true, dsp.ReedMeasure{Present: true, Dev: 4, Freq: meter.FrequencyHz(sounding, 4)}))

	tab := table(t, s)
	if tab.Transpose != 3 {
		t.Fatalf("table transpose = %d, want 3", tab.Transpose)
	}
	row := tab.Rows[0]
	if row.Note != 69 {
		t.Fatalf("row = note %d, want the A4 key", row.Note)
	}
	if got, want := row.Reeds[0].CurrHz, meter.FrequencyHz(sounding, 4)-sounding; !near(got, want) {
		t.Fatalf("CurrHz = %v, want %v at the sounding C5", got, want)
	}

	// A hand edit is converted at the sounding pitch too: 2 Hz at C5 is 2/(523.25*0.0006) units.
	tab, err := s.EditReed(0, 0, 2, "hz")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := tab.Rows[0].Reeds[0].Curr, 2/(sounding*0.0006); !near(got, want) {
		t.Fatalf("edited Curr = %v, want %v", got, want)
	}
}

// Latest names the take Undo removes next: the newest capture, even when it replaced a row in place.
func TestTheTableNamesTheLatestTake(t *testing.T) {
	s, sessions := services(t)
	open(t, sessions, 2)
	if got := table(t, s).Latest; got != -1 {
		t.Fatalf("latest on an empty pass = %d, want -1", got)
	}
	s.SetArmed(true)
	lock(s, 60, -2, 2)
	lock(s, 62, -2, 2)
	lock(s, 60, -3, 3)

	tab := table(t, s)
	if tab.Latest != 0 {
		t.Fatalf("latest = %d, want 0: note 60 was replayed last", tab.Latest)
	}
	tab, err := s.Undo()
	if err != nil {
		t.Fatal(err)
	}
	if len(tab.Rows) != 1 || tab.Rows[0].Note != 62 || tab.Latest != 0 {
		t.Fatalf("after undo rows %+v latest %d, want 62 alone", tab.Rows, tab.Latest)
	}
}

// A reading outside the curve has no goal: the table says so rather than printing a zero goal.
func TestARowPastTheCurveHasNoGoal(t *testing.T) {
	s, sessions := services(t)
	open(t, sessions, 2)
	for _, n := range []int{60, 64} {
		for reed, dev := range []float64{-5, 5} {
			if err := sessions.SetAnchor(n, reed, dev, "reference"); err != nil {
				t.Fatal(err)
			}
		}
	}
	s.SetArmed(true)
	lock(s, 62, -4, 4)
	lock(s, 70, -4, 4)

	tab := table(t, s)
	if r := tab.Rows[0]; !r.Reeds[0].HasGoal || !r.Beats[0].HasGoal {
		t.Fatalf("row 62 = %+v, want goals inside the curve", r)
	}
	if r := tab.Rows[1]; r.Reeds[0].HasGoal || r.Reeds[0].Goal != 0 || r.Beats[0].HasGoal {
		t.Fatalf("row 70 = %+v, want no goal past the curve", r)
	}
}
