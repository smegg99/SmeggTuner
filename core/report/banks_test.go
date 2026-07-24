package report

import (
	"strings"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/tuning"
)

func bank(b ...session.Bank) []session.Bank { return b }

func lmmm() session.Instrument {
	return session.Instrument{
		ReedCount: 4,
		Banks:     bank(session.BankL, session.BankM1, session.BankM2, session.BankM3),
		Registers: []session.Register{
			{Name: "LMMM", Banks: bank(session.BankL, session.BankM1, session.BankM2, session.BankM3)},
			{Name: "MMM", Banks: bank(session.BankM1, session.BankM2, session.BankM3)},
			{Name: "M2", Banks: bank(session.BankM2)},
			{Name: "L", Banks: bank(session.BankL)},
		},
	}
}

func played(s *session.Session, note tuning.Note, register string, cents ...float64) {
	t := reeds(note, cents...)
	t.Register = register
	s.UpsertTake(t)
}

func report(t *testing.T, s *session.Session) *Report {
	t.Helper()
	rep, err := Build(s, Options{Now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	return rep
}

func TestAReadingIsPrintedUnderTheRankItCameFrom(t *testing.T) {
	s := session.New("Bench", lmmm(), passA4)
	played(s, 69, "M2", 7)

	rep := report(t, s)
	if len(rep.Rows) != 1 {
		t.Fatalf("%d rows", len(rep.Rows))
	}

	cells := rep.Rows[0].Reeds
	if len(cells) != 4 {
		t.Fatalf("%d columns", len(cells))
	}
	for i, c := range cells {
		want := rep.Banks[i] == session.BankM2
		if c.Present != want {
			t.Fatalf("column %s is present=%v, want %v", rep.Head(i), c.Present, want)
		}
	}
	if got := cells[2]; got.Curr < 6 || got.Curr > 8 {
		t.Fatalf("the M2 column reads %v cents, want the reading that was taken", got.Curr)
	}
}

func TestTheColumnsAreCalledWhatTheSwitchesAre(t *testing.T) {
	s := session.New("Bench", lmmm(), passA4)
	played(s, 69, "LMMM", -4, -2, 2, 4)

	rep := report(t, s)
	for i, want := range []string{"L", "M1", "M2", "M3"} {
		if got := rep.Head(i); got != want {
			t.Fatalf("column %d is called %q, want %q", i, got, want)
		}
	}
}

func TestAnInstrumentWithNoRanksStillPrints(t *testing.T) {
	s := session.New("Bench", session.Instrument{ReedCount: 3}, passA4)
	s.UpsertTake(reeds(69, -4, 0, 4))

	rep := report(t, s)
	if len(rep.Banks) != 0 {
		t.Fatalf("banks = %v, want none", rep.Banks)
	}
	if got := rep.Head(0); got != "Reed 1" {
		t.Fatalf("column 0 is called %q", got)
	}
	if len(rep.Rows[0].Reeds) != 3 {
		t.Fatal("the reading was dropped")
	}
}

func TestAReadingThatDoesNotFitItsRegisterIsNotFiledUnderIt(t *testing.T) {
	s := session.New("Bench", lmmm(), passA4)
	played(s, 69, "M2", -4, 0, 4) // three reeds out of a single rank

	rep := report(t, s)
	cells := rep.Rows[0].Reeds
	for i := range 3 {
		if !cells[i].Present {
			t.Fatalf("the reading was dropped from column %s", rep.Head(i))
		}
	}
	if cells[3].Present {
		t.Fatal("a fourth reed appeared out of nowhere")
	}
}

func TestTheCardNamesTheSwitchOnlyWhenItChanges(t *testing.T) {
	one := session.New("Bench", lmmm(), passA4)
	played(one, 69, "LMMM", -4, -2, 2, 4)
	played(one, 71, "LMMM", -4, -2, 2, 4)
	if rep := report(t, one); rep.MultiRegister {
		t.Fatal("a session swept on one register prints a register column on every row")
	}

	two := session.New("Bench", lmmm(), passA4)
	played(two, 69, "LMMM", -4, -2, 2, 4)
	played(two, 69, "M2", 7)

	rep := report(t, two)
	if !rep.MultiRegister {
		t.Fatal("a session on two registers does not say which row is which")
	}
	if len(rep.Rows) != 2 {
		t.Fatalf("%d rows, want one per register", len(rep.Rows))
	}
	if rep.Rows[0].Register != "LMMM" || rep.Rows[1].Register != "M2" {
		t.Fatalf("rows are %q and %q", rep.Rows[0].Register, rep.Rows[1].Register)
	}
}

func TestTheRenderedCardPrintsTheRankNames(t *testing.T) {
	s := session.New("Bench", lmmm(), passA4)
	played(s, 69, "LMMM", -4, -2, 2, 4)

	var b strings.Builder
	if err := HTML(&b, report(t, s)); err != nil {
		t.Fatal(err)
	}
	html := b.String()

	for _, want := range []string{">L<", ">M1<", ">M2<", ">M3<"} {
		if !strings.Contains(html, want) {
			t.Fatalf("the printed card has no %q column", strings.Trim(want, "<>"))
		}
	}
	if strings.Contains(html, "Reed 1") {
		t.Fatal("the printed card still numbers a column it can name")
	}
}
