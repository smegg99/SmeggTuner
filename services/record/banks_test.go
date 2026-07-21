package record

import (
	"slices"
	"testing"
	"time"

	coresession "smegg.me/smeggtuner/core/session"
)

// A row's reeds land in the columns of the register they were played on, low to high.
func TestAReedLandsInTheColumnOfTheRankItCameFrom(t *testing.T) {
	rec, sessions := services(t)
	openInstrument(t, sessions, morino())

	if err := sessions.SetRegister("M1M3"); err != nil {
		t.Fatal(err)
	}
	if err := sessions.UpsertTake(coresession.Take{
		Note:  69,
		Reeds: reeds(69, -4, 4),
	}); err != nil {
		t.Fatal(err)
	}

	tab := table(t, rec)
	if len(tab.Rows) != 1 {
		t.Fatalf("%d rows", len(tab.Rows))
	}
	row := tab.Rows[0]
	if row.Register != "M1M3" {
		t.Fatalf("register = %q", row.Register)
	}
	if !slices.Equal(row.Banks, banks(coresession.BankM1, coresession.BankM3)) {
		t.Fatalf("the two reeds landed in %v, want M1 and M3", row.Banks)
	}
	if !slices.Equal(tab.Banks, morino().Banks) {
		t.Fatalf("the table prints columns %v", tab.Banks)
	}
}

// One note on two registers is two rows: MMM and M2 read different reeds, and collapsing them would lose one.
func TestOneNoteOnTwoRegistersIsTwoRows(t *testing.T) {
	rec, sessions := services(t)
	openInstrument(t, sessions, morino())

	if err := sessions.SetRegister("MMM"); err != nil {
		t.Fatal(err)
	}
	if err := sessions.UpsertTake(coresession.Take{Note: 60, At: time.Now(), Reeds: reeds(60, -4, 0, 4)}); err != nil {
		t.Fatal(err)
	}
	if err := sessions.SetRegister("M2"); err != nil {
		t.Fatal(err)
	}
	if err := sessions.UpsertTake(coresession.Take{Note: 60, At: time.Now(), Reeds: reeds(60, 9)}); err != nil {
		t.Fatal(err)
	}

	tab := table(t, rec)
	if len(tab.Rows) != 2 {
		t.Fatalf("%d rows, want one per register", len(tab.Rows))
	}
	if tab.Rows[0].Register != "M2" || tab.Rows[1].Register != "MMM" {
		t.Fatalf("rows are %q and %q", tab.Rows[0].Register, tab.Rows[1].Register)
	}
	if len(tab.Rows[0].Reeds) != 1 || len(tab.Rows[1].Reeds) != 3 {
		t.Fatal("the two rows do not hold what their registers sound")
	}
}

func TestASecondTakeOfTheSameVoiceIsStillOneRow(t *testing.T) {
	rec, sessions := services(t)
	openInstrument(t, sessions, morino())

	if err := sessions.SetRegister("M2"); err != nil {
		t.Fatal(err)
	}
	if err := sessions.UpsertTake(coresession.Take{Note: 60, At: time.Now(), Reeds: reeds(60, 9)}); err != nil {
		t.Fatal(err)
	}
	if err := sessions.UpsertTake(coresession.Take{Note: 60, At: time.Now(), Reeds: reeds(60, 2)}); err != nil {
		t.Fatal(err)
	}

	tab := table(t, rec)
	// The second reading replaced the first, so one row.
	if len(tab.Rows) != 1 {
		t.Fatalf("rows = %+v", tab.Rows)
	}
	if got := tab.Rows[0].Reeds[0].Curr; got < 1 || got > 3 {
		t.Fatalf("the row shows %v cents, want the newest reading", got)
	}
}

// A reading with more reeds than its register sounds gets no banks; the table numbers its reeds instead.
func TestAReadingThatDoesNotFitItsRegisterIsNotGivenColumns(t *testing.T) {
	rec, sessions := services(t)
	openInstrument(t, sessions, morino())

	if err := sessions.SetRegister("M1M3"); err != nil {
		t.Fatal(err)
	}
	// Three reeds out of a switch that sounds two.
	if err := sessions.UpsertTake(coresession.Take{Note: 69, At: time.Now(), Reeds: reeds(69, -4, 0, 4)}); err != nil {
		t.Fatal(err)
	}

	row := table(t, rec).Rows[0]
	if row.Banks != nil {
		t.Fatalf("three reeds were filed into %v", row.Banks)
	}
	if len(row.Reeds) != 3 {
		t.Fatal("the reading itself was dropped; it should be shown, just not in columns")
	}
}

// A take that names no register is still printed, numbered rather than in columns.
func TestATakeThatNamesNoRegisterGetsNoColumns(t *testing.T) {
	rec, sessions := services(t)
	openInstrument(t, sessions, coresession.Instrument{ReedCount: 3})

	if err := sessions.UpsertTake(coresession.Take{Note: 69, At: time.Now(), Reeds: reeds(69, -4, 0, 4)}); err != nil {
		t.Fatal(err)
	}

	tab := table(t, rec)
	if tab.Banks != nil {
		t.Fatalf("an instrument nobody described prints columns %v", tab.Banks)
	}
	if tab.Rows[0].Banks != nil {
		t.Fatalf("row banks = %v", tab.Rows[0].Banks)
	}
	if len(tab.Rows[0].Reeds) != 3 {
		t.Fatal("the reading was dropped")
	}
}

// An edit aims at a voice, not a note: with one note on two registers, the edit must hit the aimed row, not whichever was captured last.
func TestAnEditLandsOnTheRowItWasAimedAt(t *testing.T) {
	rec, sessions := services(t)
	openInstrument(t, sessions, morino())

	if err := sessions.SetRegister("MMM"); err != nil {
		t.Fatal(err)
	}
	if err := sessions.UpsertTake(coresession.Take{Note: 60, At: time.Now(), Reeds: reeds(60, -4, 0, 4)}); err != nil {
		t.Fatal(err)
	}
	if err := sessions.SetRegister("M2"); err != nil {
		t.Fatal(err)
	}
	if err := sessions.UpsertTake(coresession.Take{Note: 60, At: time.Now(), Reeds: reeds(60, 9)}); err != nil {
		t.Fatal(err)
	}

	// Row 0 is the M2 one (it sorts before MMM).
	tab := table(t, rec)
	row := tab.Rows[0]
	if row.Register != "M2" {
		t.Fatalf("row 0 is %q", row.Register)
	}
	out, err := rec.EditReed(row.Take, 0, 0, "cent")
	if err != nil {
		t.Fatal(err)
	}

	if got := out.Rows[0].Reeds[0].Curr; got < -0.5 || got > 0.5 {
		t.Fatalf("the M2 row now reads %v cents", got)
	}
	if len(out.Rows[1].Reeds) != 3 {
		t.Fatal("the MMM row lost its reeds")
	}
	if got := out.Rows[1].Reeds[0].Curr; got > -3 {
		t.Fatalf("the edit landed on the MMM row instead: %v cents", got)
	}
}
