package report

import (
	"strings"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/session"
)

func bassSession() *session.Session {
	s := &session.Session{
		Name: "Bass pass",
		A4:   440,
		Instrument: session.Instrument{ReedCount: 2,
			Banks: []session.Bank{session.BankM1, session.BankM2},
			Registers: []session.Register{
				{Name: "MM", Banks: []session.Bank{session.BankM1, session.BankM2}},
			},
			BassReeds:     5,
			BassRegisters: []session.BassRegister{{Name: "Soft", Feet: []int{16, 8}}},
		},
		Updated: time.Unix(100, 0),
	}
	s.Takes = []session.Take{
		// A treble take, so both tables print.
		{Note: 69, Register: "MM", At: time.Unix(1, 0),
			Reeds: []dsp.ReedMeasure{{Freq: 439.5}, {Freq: 441.5}}},
		// The whole ladder on the F button.
		{Note: 29, Bass: true, At: time.Unix(2, 0), Reeds: []dsp.ReedMeasure{
			{Freq: 43.7}, {Freq: 87.5, Octave: 12}, {Freq: 174.8, Octave: 24},
			{Freq: 349.7, Octave: 36}, {Freq: 699.2, Octave: 48}}},
		// A partial switch: its 16' and 8' must land under those columns, not the first two.
		{Note: 45, Bass: true, Register: "Soft", At: time.Unix(3, 0), Reeds: []dsp.ReedMeasure{
			{Freq: 110.1}, {Freq: 220.3, Octave: 12}}},
	}
	return s
}

func TestReportLaysTheBassSideAsItsOwnTable(t *testing.T) {
	rep, err := Build(bassSession(), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Rows) != 1 {
		t.Fatalf("treble table: want the one treble take, got %d rows", len(rep.Rows))
	}
	if rep.Bass == nil {
		t.Fatal("a pass with bass takes must carry the bass table")
	}
	if got := rep.Bass.Feet; len(got) != 5 || got[0] != 32 || got[4] != 2 {
		t.Fatalf("bass columns = %v, want the machine's 32..2 ladder", got)
	}
	if rep.Bass.Head(0) != "32'" {
		t.Errorf("bass column heading = %q, want the foot", rep.Bass.Head(0))
	}
	if len(rep.Bass.Rows) != 2 {
		t.Fatalf("bass table: want 2 rows, got %d", len(rep.Bass.Rows))
	}

	// The full ladder fills every column; the Soft switch's two reeds land under 16' and 8'.
	full := rep.Bass.Rows[0]
	for i, c := range full.Reeds {
		if !c.Present {
			t.Errorf("full-ladder row: column %d empty", i)
		}
	}
	soft := rep.Bass.Rows[1]
	present := []bool{}
	for _, c := range soft.Reeds {
		present = append(present, c.Present)
	}
	want := []bool{false, true, true, false, false}
	for i := range want {
		if present[i] != want[i] {
			t.Fatalf("Soft row columns present = %v, want %v (16' and 8' only)", present, want)
		}
	}
}

func TestBassTableRendersAndExports(t *testing.T) {
	rep, err := Build(bassSession(), Options{})
	if err != nil {
		t.Fatal(err)
	}
	var html strings.Builder
	if err := HTML(&html, rep); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(html.String(), "32&#39;") && !strings.Contains(html.String(), "32'") {
		t.Error("the printed sheet must head the bass table with the feet")
	}

	var csv strings.Builder
	if err := CSV(&csv, rep); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(csv.String(), "bass_32ft_curr_cents") {
		t.Error("the spreadsheet must carry the bass block keyed by foot")
	}
}
