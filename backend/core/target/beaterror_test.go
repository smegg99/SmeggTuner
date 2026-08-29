package target

import (
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// The beat goal is not typed; it is what the two slots' goals say about the pair.
func TestBeatErrorsAgainstCurve(t *testing.T) {
	c := musette(t) // at note 72 the goals are -9 / 0 / +9
	m := measure(72, 440, -6.2, 0.5, 11.0)
	got := BeatErrors(m, c, 440, 0, nil)

	if len(got) != 2 {
		t.Fatalf("three slots beat in two adjacent pairs, got %d: %+v", len(got), got)
	}
	ref := tuning.Note(72).Freq(440)
	want := []struct {
		pair       string
		low, high  int
		curr, goal float64
	}{
		{"1-2", 0, 1, 0.5 - -6.2, 9}, // the flat reed against the reed at pitch
		{"2-3", 1, 2, 11.0 - 0.5, 9},
	}
	for i, w := range want {
		b := got[i]
		if b.Pair != w.pair || b.Low != w.low || b.High != w.high {
			t.Fatalf("row %d is %q (slots %d,%d)", i, b.Pair, b.Low, b.High)
		}
		almost(t, b.Curr, w.curr, 1e-9, b.Pair+" beat measured")
		almost(t, b.Goal, w.goal, 1e-9, b.Pair+" beat asked for")
		almost(t, b.Error, w.curr-w.goal, 1e-9, b.Pair+" beat error")
		almost(t, b.GoalHz, w.goal*ref*0.0006, 1e-9, b.Pair+" goal in Hz")
		almost(t, b.ErrorHz, b.CurrHz-b.GoalHz, 1e-12, b.Pair+" error in Hz")
		if (b.Error < 0) != (b.ErrorHz < 0) {
			t.Fatalf("%s: %v units and %v Hz disagree on which way the pair is out",
				b.Pair, b.Error, b.ErrorHz)
		}
	}
}

// A tremolo pair need not be within a unit, so the beat window is its own number.
func TestBeatToleranceIsNotTheReedTolerance(t *testing.T) {
	c := musette(t)
	m := measure(72, 440, -6.2, 0.5, 11.0) // pair 1-2 lands 2.3 units out

	for _, b := range BeatErrors(m, c, 440, 0, nil) {
		if !b.InTol {
			t.Fatalf("%s is %.2f units out and inside the %v unit beat window",
				b.Pair, b.Error, DefaultBeatTolerance)
		}
	}
	if BeatErrors(m, c, 440, DefaultTolerance, nil)[0].InTol {
		t.Fatal("2.3 units is not inside a one unit window")
	}
}

// The beat is the engine's measured rate, never the difference of the two reed readings.
func TestTheBeatIsTheMeasuredRate(t *testing.T) {
	c := musette(t)
	ref := tuning.Note(72).Freq(440)
	m := measure(72, 440, -1, 1)
	m.Beats = []dsp.BeatMeasure{{Pair: "1-2", Present: true, Hz: 2.5}}

	got := BeatErrors(m, c, 440, 0, nil)
	if len(got) != 1 {
		t.Fatalf("want the one measured beat, got %+v", got)
	}
	b := got[0]
	almost(t, b.CurrHz, 2.5, 1e-12, "the beat the engine heard")
	almost(t, b.Curr, meter.DeviationUnits(ref+2.5, ref), 1e-9, "the same beat in units")
	almost(t, b.Goal, 9, 1e-9, "the curve still asks for its nine units")
}

// Nothing to aim at: the beat is its own error, exactly as a reed's is.
func TestBeatErrorsWithoutCurve(t *testing.T) {
	m := measure(tuning.NoteA4, 440, -6, 0, 6)
	for _, b := range BeatErrors(m, nil, 440, 0, nil) {
		if b.Goal != 0 || b.GoalHz != 0 || b.HasGoal {
			t.Fatalf("%s: goal %v with no curve", b.Pair, b.Goal)
		}
		almost(t, b.Error, b.Curr, 1e-12, "the error is the beat")
		almost(t, b.ErrorHz, b.CurrHz, 1e-12, "in Hz too")
	}
}

func TestBeatErrorsReedCounts(t *testing.T) {
	if got := BeatErrors(measure(tuning.NoteA4, 440, 2), nil, 440, 0, nil); len(got) != 0 {
		t.Fatalf("one slot gave %d beats", len(got))
	}
	if got := BeatErrors(measure(60, 440, 0, 1, 2, 3, 4), nil, 440, 0, nil); len(got) != 4 {
		t.Fatalf("five slots gave %d beats, want the 4 adjacent pairs", len(got))
	}
	if got := BeatErrors(Reading{}, nil, 440, 0, nil); len(got) != 0 {
		t.Fatalf("a heartbeat gave %d beats", len(got))
	}
}

// A slot that did not sound takes both its pairs with it.
func TestAMissingSlotHasNoBeats(t *testing.T) {
	if got := BeatErrors(measure(tuning.NoteA4, 440, -6, gone, 6), musette(t), 440, 0, nil); len(got) != 0 {
		t.Fatalf("the pairs around a silent slot printed: %+v", got)
	}
	got := BeatErrors(measure(tuning.NoteA4, 440, -6, 0, gone), musette(t), 440, 0, nil)
	if len(got) != 1 || got[0].Pair != "1-2" {
		t.Fatalf("want only 1-2, got %+v", got)
	}
}

// A partial register's two slots sit in columns that are not neighbours on the card; the goal
// is read there, and the rate converts at the upper slot's pitch, where a pair two octaves apart
// beats at the difference of its deviations.
func TestAPairIsScoredAgainstItsOwnColumns(t *testing.T) {
	const a4 = 440
	c := NewCurve("LMMH", 4)
	for col, v := range []float64{-6, 0, 6, 12} {
		mustSet(t, c, tuning.NoteA4, col, v)
	}
	low, high := tuning.NoteA4.Freq(a4)/2, tuning.NoteA4.Freq(a4)*2
	m := Reading{Note: tuning.NoteA4, Reeds: []dsp.ReedMeasure{
		{Present: true, Dev: -6, Octave: -12, Freq: meter.FrequencyHz(low, -6)},
		{Present: true, Dev: 12, Octave: 12, Freq: meter.FrequencyHz(high, 12)},
	}}
	m.Beats = beats(m.Reeds)

	got := BeatErrors(m, c, a4, 0, []int{0, 3})
	if len(got) != 1 {
		t.Fatalf("%d beats, want 1", len(got))
	}
	almost(t, got[0].Goal, 18, 1e-9, "the H goal minus the L goal, not M1 minus L")
	almost(t, got[0].GoalHz, 18*high*0.0006, 1e-9, "converted at the 4' slot's pitch")
	almost(t, got[0].Error, 0, 1e-6, "both reeds on their goals")
}

// A 16' under an 8', both exactly on their goals, beats exactly as the goals ask: the pair is read at the 8' pitch.
func TestAnOctavePairOnGoalIsInTune(t *testing.T) {
	c := musette(t) // at note 72 the first two columns ask for -9 and 0
	low := meter.FrequencyHz(tuning.Note(60).Freq(440), -9)
	high := meter.FrequencyHz(tuning.Note(72).Freq(440), 0)
	r := Reading{Note: 72, Reeds: []dsp.ReedMeasure{
		{Present: true, Dev: -9, Freq: low, Octave: -12},
		{Present: true, Dev: 0, Freq: high},
	}}
	r.Beats = beats(r.Reeds)

	got := BeatErrors(r, c, 440, 0, nil)
	if len(got) != 1 {
		t.Fatalf("want one pair, got %+v", got)
	}
	almost(t, got[0].Curr, 9, 1e-6, "the octave beat in units at the 8' pitch")
	almost(t, got[0].Error, 0, 1e-6, "an octave pair on its goals")
	almost(t, got[0].ErrorHz, 0, 1e-6, "the same in Hz")
}
