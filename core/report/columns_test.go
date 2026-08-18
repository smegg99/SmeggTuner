package report

import (
	"testing"

	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/tuning"
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

// octaveTake is a reading of a register that spans octaves: each reed carries the band it was
// found in, which is how the card knows which rank it is.
func octaveTake(note tuning.Note, register string, ranks ...dsp.ReedMeasure) session.Take {
	t := session.Take{Note: note, At: at, Register: register, Reeds: ranks}
	return t
}

func rank(note tuning.Note, octave int, cents float64) dsp.ReedMeasure {
	ref := tuning.FreqAtCents(note.Freq(passA4), 0) * bandFactor(octave)
	return dsp.ReedMeasure{Freq: tuning.FreqAtCents(ref, cents), DevCents: cents, Octave: octave}
}

func bandFactor(octave int) float64 {
	switch octave {
	case -12:
		return 0.5
	case 12:
		return 2
	}
	return 1
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
	s.UpsertTake(octaveTake(69, "LH", rank(69, -12, -10), rank(69, 12, 20)))

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

// The beat between two ranks belongs in the column between them. On an LMMH card with the second
// 8' unheard, the 8'-to-4' beat spans the empty M2 column - it is not the musette beat, and
// printing it in the M1-M2 column would read as one.
func TestAnOctaveBeatDoesNotPrintInTheMusetteColumn(t *testing.T) {
	s := session.New("Bench", lmmh(), passA4)
	take := octaveTake(67, "LMMH", rank(67, -12, 0), rank(67, 0, 0), rank(67, 12, 3))
	// The engine's own naming: reeds 2 and 3, which are the 8' and the 4'.
	take.Beats = []dsp.BeatMeasure{{Pair: "2-3", Hz: 1.17, FromEnvelope: true, Depth: 0.13}}
	take.ReedsMerged = true // a band short of its ranks: the beat is the reading
	s.UpsertTake(take)

	r := sheet(t, s)
	var musette, octave BeatCell
	for i, p := range r.Pairs {
		switch {
		case p.Low == 2 && p.High == 3:
			musette = row(t, r, 67).Beats[i]
		case p.Low == 2 && p.High == 4:
			octave = row(t, r, 67).Beats[i]
		}
	}
	if musette.Present {
		t.Errorf("the M1-M2 musette column prints %.2f Hz: that beat is between the 8' and the 4'", musette.CurrHz)
	}
	if !octave.Present {
		t.Fatalf("no M1-H column: the 8'-to-4' beat has nowhere to print. Columns: %+v", r.Pairs)
	}
	if !near(octave.CurrHz, 1.17, 0.05) {
		t.Errorf("the M1-H column reads %.3f Hz, want 1.17", octave.CurrHz)
	}
}

// A band that heard two ranks but could only split one reports the pair off its envelope. That
// beat's upper index points at whatever reed came next - which is the band above's, or none - so
// only the beat's own octaves say it never left its band. It belongs in that band's own column.
func TestAnUnsplitMusettePairStaysInItsOwnColumn(t *testing.T) {
	s := session.New("Bench", lmmh(), passA4)
	// L, one of two 8' ranks, and the 4': the M band heard a beat it could not split.
	take := octaveTake(67, "LMMH", rank(67, -12, 0), rank(67, 0, 0), rank(67, 12, 3))
	take.ReedsMerged = true
	take.Beats = []dsp.BeatMeasure{
		{Pair: "2-3", Hz: 2.4, HighUnheard: true, FromEnvelope: true, Depth: 0.7},
		{Pair: "2-3", Hz: 1.17, FromEnvelope: true, Depth: 0.13},
	}
	s.UpsertTake(take)

	r := sheet(t, s)
	cells := map[[2]int]BeatCell{}
	for i, p := range r.Pairs {
		cells[[2]int{p.Low, p.High}] = row(t, r, 67).Beats[i]
	}
	musette, ok := cells[[2]int{2, 3}]
	if !ok || !musette.Present {
		t.Fatalf("no reading in the M1-M2 column: the unsplit pair was filed elsewhere. Columns: %+v", r.Pairs)
	}
	if !near(musette.CurrHz, 2.4, 0.05) {
		t.Errorf("the M1-M2 column reads %.3f Hz, want 2.40", musette.CurrHz)
	}
	octave, ok := cells[[2]int{2, 4}]
	if !ok || !octave.Present {
		t.Fatalf("no reading in the M1-H column: the octave beat was lost. Columns: %+v", r.Pairs)
	}
	if !near(octave.CurrHz, 1.17, 0.05) {
		t.Errorf("the M1-H column reads %.3f Hz, want 1.17", octave.CurrHz)
	}
}
