package target

import (
	"math"
	"testing"

	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/tuning"
)

// The beat goal is not typed; it is what the two reeds' goals say about the pair.
func TestBeatErrorsAgainstCurve(t *testing.T) {
	c := musette(t) // at note 72 the goals are -9 / 0 / +9
	m := measure(72, 440, -6.2, 0.5, 11.0)
	got := BeatErrors(m, c, 440, 0, nil)

	if len(got) != 3 {
		t.Fatalf("three reeds beat in three pairs, got %d: %+v", len(got), got)
	}
	want := []struct {
		pair       string
		low, high  int
		curr, goal float64
	}{
		{"1-2", 0, 1, 0.5 - -6.2, 9},   // the flat reed against the reed at pitch
		{"1-3", 0, 2, 11.0 - -6.2, 18}, // and against the sharp one: the full spread
		{"2-3", 1, 2, 11.0 - 0.5, 9},
	}
	for i, w := range want {
		b := got[i]
		if b.Pair != w.pair || b.Low != w.low || b.High != w.high {
			t.Fatalf("row %d is %q (reeds %d,%d)", i, b.Pair, b.Low, b.High)
		}
		almost(t, b.Curr, w.curr, 1e-9, b.Pair+" beat measured")
		almost(t, b.Goal, w.goal, 1e-9, b.Pair+" beat asked for")
		almost(t, b.Error, w.curr-w.goal, 1e-9, b.Pair+" beat error")
		if b.FromEnvelope {
			t.Fatalf("%s came off separated peaks, not the envelope", b.Pair)
		}

		almost(t, b.GoalHz, HzFromCents(72, w.goal, 440), 1e-9, b.Pair+" goal in Hz")
		almost(t, b.ErrorHz, b.CurrHz-b.GoalHz, 1e-12, b.Pair+" error in Hz")
		if (b.Error < 0) != (b.ErrorHz < 0) {
			t.Fatalf("%s: %v cents and %v Hz disagree on which way the pair is out",
				b.Pair, b.Error, b.ErrorHz)
		}
	}
}

// A tremolo pair need not be within a cent, so the beat window is its own number.
func TestBeatToleranceIsNotTheReedTolerance(t *testing.T) {
	c := musette(t)
	m := measure(72, 440, -6.2, 0.5, 11.0) // pair 1-2 lands 2.3 cents out

	for _, b := range BeatErrors(m, c, 440, 0, nil) {
		if !b.InTol {
			t.Fatalf("%s is %.2f cents out and inside the %v cent beat window",
				b.Pair, b.Error, DefaultBeatTolerance)
		}
	}
	if BeatErrors(m, c, 440, DefaultTolerance, nil)[0].InTol {
		t.Fatal("2.3 cents is not inside a one cent window")
	}
}

// The reeds do not pull apart, so the peaks are lobes of a merged pair; only the envelope beat survives.
func TestBeatErrorsMergedPair(t *testing.T) {
	c := musette(t)
	ref := tuning.Note(72).Freq(440)
	m := dsp.Measurement{
		Note:       72,
		ScalePitch: ref,
		// Two lobes a couple of cents apart: believing them reports 2 cents when it beats at eight.
		Reeds:          []dsp.ReedMeasure{{Freq: tuning.FreqAtCents(ref, -1)}, {Freq: tuning.FreqAtCents(ref, 1)}},
		ReedsSeparated: false,
		Beats: []dsp.BeatMeasure{
			{Pair: "1-2", Hz: 2.5, Cents: tuning.Cents(ref+2.5, ref), FromEnvelope: true, Depth: 0.7},
		},
	}

	got := BeatErrors(m, c, 440, 0, nil)
	if len(got) != 1 {
		t.Fatalf("only the measured beat may be reported, got %+v", got)
	}
	b := got[0]
	if b.Pair != "1-2" || !b.FromEnvelope {
		t.Fatalf("row is %+v", b)
	}
	almost(t, b.CurrHz, 2.5, 1e-9, "the beat the engine actually heard")
	almost(t, b.Curr, tuning.Cents(ref+2.5, ref), 1e-9, "the same beat in cents")
	if b.Curr < 5 {
		t.Fatalf("the lobes were believed: %v cents", b.Curr)
	}
	almost(t, b.Goal, 9, 1e-9, "the curve still asks for its nine cents")
}

// Nothing to aim at: the beat is its own error, exactly as a reed's is.
func TestBeatErrorsWithoutCurve(t *testing.T) {
	m := measure(tuning.NoteA4, 440, -6, 0, 6)
	for _, b := range BeatErrors(m, nil, 440, 0, nil) {
		if b.Goal != 0 || b.GoalHz != 0 {
			t.Fatalf("%s: goal %v with no curve", b.Pair, b.Goal)
		}
		almost(t, b.Error, b.Curr, 1e-12, "the error is the beat")
		almost(t, b.ErrorHz, b.CurrHz, 1e-12, "in Hz too")
	}
}

func TestBeatErrorsReedCounts(t *testing.T) {
	if got := BeatErrors(measure(tuning.NoteA4, 440, 2), nil, 440, 0, nil); len(got) != 0 {
		t.Fatalf("one reed gave %d beats", len(got))
	}
	if got := BeatErrors(measure(60, 440, 0, 1, 2, 3, 4), nil, 440, 0, nil); len(got) != 10 {
		t.Fatalf("five reeds gave %d beats, want 10", len(got))
	}
	if got := BeatErrors(dsp.Measurement{State: dsp.StateTooQuiet}, nil, 440, 0, nil); len(got) != 0 {
		t.Fatalf("a heartbeat gave %d beats", len(got))
	}
}

// A band that heard a pair but could not split it names only its lower rank; the upper index points
// at whatever reed came next. The goal for such a pair is between the band's OWN two columns, and
// the rate converts at the band's own octave - reading either off the reed the index landed on puts
// an 8'-to-8' beat against a 4's goal, an octave up.
func TestAnUnsplitPairIsScoredAgainstItsOwnBandsColumns(t *testing.T) {
	// L, one 8' of two, and a 4': the M band could not split its pair.
	m := dsp.Measurement{
		Note:       tuning.NoteA4,
		ScalePitch: tuning.NoteA4.Freq(440),
		Reeds: []dsp.ReedMeasure{
			{Freq: 220, Octave: -12}, {Freq: 440, Octave: 0}, {Freq: 880, Octave: 12},
		},
		Beats: []dsp.BeatMeasure{{Pair: "2-3", Hz: 2, HighUnheard: true, FromEnvelope: true}},
	}
	c := NewCurve("LMMH", 4)
	for col, v := range []float64{-6, 0, 6, 12} {
		if err := c.Set(tuning.NoteA4, col, v, 440); err != nil {
			t.Fatal(err)
		}
	}

	got := BeatErrors(m, c, 440, 0, []int{0, 1, 3})
	if len(got) != 1 {
		t.Fatalf("%d beats, want 1", len(got))
	}
	b := got[0]
	if !b.HighUnheard {
		t.Fatal("the reading lost the fact that its upper rank was never heard")
	}
	// M1 to M2, not M1 to H: 6 - 0, not 12 - 0.
	if math.Abs(b.Goal-6) > 0.01 {
		t.Errorf("goal %.2f cents, want 6: the pair was scored against the 4' column", b.Goal)
	}
	// 2 Hz at 440, not at 880.
	if math.Abs(b.CurrHz-2) > 0.01 {
		t.Errorf("rate %.3f Hz, want 2.000: it converted at the wrong octave", b.CurrHz)
	}
}
