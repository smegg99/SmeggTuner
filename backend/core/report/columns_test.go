package report

import (
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// lmmh is the Bugari's card: 16', two 8' ranks and a 4'. Its sub-registers leave gaps.
func lmmh() session.Instrument {
	return session.Instrument{
		ReedCount: 4,
		Banks:     bank(session.BankL, session.BankM1, session.BankM2, session.BankH),
		Registers: []session.Register{
			{Name: "LMMH", Banks: bank(session.BankL, session.BankM1, session.BankM2, session.BankH)},
			{Name: "LH", Banks: bank(session.BankL, session.BankH)},
			{Name: "MH", Banks: bank(session.BankM1, session.BankH)},
		},
	}
}

// octaveTake is a reading of a register that spans octaves: slot i is the register's i-th bank
// in card order, each at its own octave.
func octaveTake(note tuning.Note, register string, ranks ...dsp.ReedMeasure) session.Take {
	return session.Take{Note: note, At: at, Register: register, Reeds: ranks, Beats: beatsOf(ranks)}
}

// A register that does not sound every rank leaves gaps in the card, and the goal has to be read at
// each rank's own column. Scored by position, the 16' of a 16'+4' take would be judged against the
// L goal (right, by luck) and the 4' against the M1 goal (wrong by a whole rank).
func TestAPartialRegisterIsJudgedAgainstItsOwnColumns(t *testing.T) {
	// One goal per column, all different, so a misread column is unmistakable.
	c := curve(t, 4, map[tuning.Note][]float64{69: {-10, 0, 6, 20}})

	s := session.New("Bench", lmmh(), passA4)
	s.Curve = c
	// Both ranks sitting exactly on their own goal: every Error must come out zero.
	s.UpsertTake(octaveTake(69, "LH", slot(69, -12, -10), slot(69, 12, 20)))

	r := sheet(t, s)
	cells := row(t, r, 69).Reeds
	if len(cells) != 4 {
		t.Fatalf("%d columns", len(cells))
	}
	for i, want := range []bool{true, false, false, true} {
		if cells[i].Present != want {
			t.Fatalf("column %s present=%v, want %v", r.Head(i), cells[i].Present, want)
		}
	}
	if got := cells[0]; !near(got.Goal, -10, 0.01) || !near(got.Error, 0, 0.01) {
		t.Errorf("the L column reads goal %.2f error %.2f, want -10 and 0", got.Goal, got.Error)
	}
	if got := cells[3]; !near(got.Goal, 20, 0.01) || !near(got.Error, 0, 0.01) {
		t.Errorf("the H column reads goal %.2f error %.2f, want 20 and 0: it was scored "+
			"against another rank's goal", got.Goal, got.Error)
	}
}

// A rank that did not sound takes both its beats with it: on an LMMH card with the second 8'
// silent there is no musette beat and no beat across the gap, only the 16'-to-8' one.
func TestASilentRankTakesItsBeatsWithIt(t *testing.T) {
	s := session.New("Bench", lmmh(), passA4)
	s.UpsertTake(octaveTake(67, "LMMH", slot(67, -12, 0), slot(67, 0, 0), slot(67, 0, gone), slot(67, 12, 3)))

	r := sheet(t, s)
	got := map[[2]int]BeatCell{}
	for i, p := range r.Pairs {
		got[[2]int{p.Low, p.High}] = row(t, r, 67).Beats[i]
	}
	if len(r.Pairs) != 3 {
		t.Fatalf("columns %+v, want only the card's three neighbour pairs", r.Pairs)
	}
	if !got[[2]int{1, 2}].Present {
		t.Error("the L-M1 beat did not print")
	}
	for _, pair := range [][2]int{{2, 3}, {3, 4}} {
		if got[pair].Present {
			t.Errorf("the %v column prints %.2f Hz next to a rank that did not sound", pair, got[pair].CurrHz)
		}
	}
}

// A partial register's two slots are neighbours in the take but not on the card: their beat gets
// a column of its own rather than landing in the L-M1 one.
func TestAPartialRegistersBeatGetsItsOwnColumn(t *testing.T) {
	s := session.New("Bench", lmmh(), passA4)
	s.UpsertTake(octaveTake(69, "LH", slot(69, -12, -10), slot(69, 12, 20)))

	r := sheet(t, s)
	cells := map[[2]int]BeatCell{}
	for i, p := range r.Pairs {
		cells[[2]int{p.Low, p.High}] = row(t, r, 69).Beats[i]
	}
	if cells[[2]int{1, 2}].Present {
		t.Error("the L-H beat printed in the L-M1 column")
	}
	lh, ok := cells[[2]int{1, 4}]
	if !ok || !lh.Present {
		t.Fatalf("no L-H column: the 16'-to-4' beat has nowhere to print. Columns: %+v", r.Pairs)
	}
	if !near(lh.CurrHz, s.Takes[0].Beats[0].Hz, 1e-9) {
		t.Errorf("the L-H column reads %.3f Hz, want the measured %.3f", lh.CurrHz, s.Takes[0].Beats[0].Hz)
	}
}
