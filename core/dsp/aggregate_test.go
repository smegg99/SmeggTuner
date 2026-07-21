package dsp

import (
	"math"
	"testing"

	"smegg.me/smeggtuner/core/tuning"
)

// aggTick builds a running measurement the way the engine would emit it.
func aggTick(note tuning.Note, level float32, freqs ...float64) Measurement {
	m := Measurement{
		Note:       note,
		NoteName:   note.Name(tuning.NamingCDEFGAB),
		InputLevel: level,
		State:      StateRunning,
	}
	fc := note.Freq(440)
	for _, f := range freqs {
		m.Reeds = append(m.Reeds, ReedMeasure{Freq: f, DevCents: tuning.Cents(f, fc)})
	}
	return m
}

func TestAggregateRejectsUnusableInput(t *testing.T) {
	if _, ok := Aggregate(nil, 440, 0.5); ok {
		t.Fatal("empty input must not aggregate")
	}
	init := Measurement{State: StateInitializing, InputLevel: 0.2}
	if _, ok := Aggregate([]Measurement{init, init}, 440, 0.5); ok {
		t.Fatal("input without running ticks must not aggregate")
	}
	noNote := Measurement{State: StateRunning, InputLevel: 0.2}
	if _, ok := Aggregate([]Measurement{noNote}, 440, 0.5); ok {
		t.Fatal("running ticks without a note must not aggregate")
	}
}

func TestAggregateMajorityNoteOverNoisyHeadTail(t *testing.T) {
	// silent head/tail where detection latched onto room noise: low level, wrong note.
	noise := aggTick(tuning.Note(35), 0.004, 118.7)
	ms := []Measurement{noise, noise, noise}
	for i := 0; i < 8; i++ {
		ms = append(ms, aggTick(tuning.NoteA4, 0.2, 441.5))
	}
	// stray detection at sustained level loses the majority vote
	ms = append(ms, aggTick(tuning.Note(67), 0.15, 392.1))
	// non-running tick must not raise the level threshold
	loud := aggTick(tuning.Note(35), 0.9, 118.7)
	loud.State = StateTooLoud
	ms = append(ms, loud, noise, noise)

	m, ok := Aggregate(ms, 440, 0.5)
	if !ok {
		t.Fatal("expected ok")
	}
	if m.Note != tuning.NoteA4 || m.NoteName != "A4" {
		t.Fatalf("note = %s (%d) want A4", m.NoteName, m.Note)
	}
	if m.State != StateRunning {
		t.Errorf("state = %s want %s", m.State, StateRunning)
	}
	if len(m.Reeds) != 1 {
		t.Fatalf("reeds = %+v want one", m.Reeds)
	}
	if math.Abs(m.Reeds[0].Freq-441.5) > 1e-9 {
		t.Errorf("freq = %v want 441.5", m.Reeds[0].Freq)
	}
}

func TestAggregateMedianReedsAndDevCents(t *testing.T) {
	a4 := 442.0
	ms := []Measurement{
		aggTick(tuning.NoteA4, 0.3, 441.2, 443.9),
		aggTick(tuning.NoteA4, 0.3, 441.5, 444.1),
		aggTick(tuning.NoteA4, 0.3, 441.4, 444.0),
		aggTick(tuning.NoteA4, 0.3, 447.0, 451.0), // transient outlier tick
		aggTick(tuning.NoteA4, 0.3, 441.5),        // dropout: minority reed count
	}
	m, ok := Aggregate(ms, a4, 0.5)
	if !ok {
		t.Fatal("expected ok")
	}
	if len(m.Reeds) != 2 {
		t.Fatalf("reeds = %+v want two (majority reed count)", m.Reeds)
	}
	// even-sized median averages the middle pair, so the single outlier moves nothing
	want := []float64{(441.4 + 441.5) / 2, (444.0 + 444.1) / 2}
	for i, w := range want {
		if math.Abs(m.Reeds[i].Freq-w) > 1e-9 {
			t.Errorf("reed %d freq = %v want %v", i+1, m.Reeds[i].Freq, w)
		}
		dev := tuning.Cents(w, tuning.NoteA4.Freq(a4))
		if math.Abs(m.Reeds[i].DevCents-dev) > 1e-9 {
			t.Errorf("reed %d devCents = %v want %v (vs a4=%v)", i+1, m.Reeds[i].DevCents, dev, a4)
		}
	}
}

func TestAggregateBeats(t *testing.T) {
	fc := tuning.NoteA4.Freq(440)
	mk := func(hz float64, env bool) Measurement {
		m := aggTick(tuning.NoteA4, 0.3, 439.0, 439.0+hz)
		m.Beats = []BeatMeasure{{
			Pair: "1-2", Hz: hz,
			Cents:        tuning.Cents(fc+hz, fc),
			FromEnvelope: env,
		}}
		return m
	}
	ms := []Measurement{mk(2.5, false), mk(2.7, false), mk(2.6, true)}
	m, ok := Aggregate(ms, 440, 0.5)
	if !ok {
		t.Fatal("expected ok")
	}
	if len(m.Beats) != 1 {
		t.Fatalf("beats = %+v want one pair", m.Beats)
	}
	b := m.Beats[0]
	if b.Pair != "1-2" {
		t.Errorf("pair = %s want 1-2", b.Pair)
	}
	if math.Abs(b.Hz-2.6) > 1e-9 {
		t.Errorf("beat hz = %v want 2.6 (median)", b.Hz)
	}
	if b.FromEnvelope {
		t.Error("fromEnvelope must follow the majority (2 of 3 spectral)")
	}
	if want := tuning.Cents(fc+2.6, fc); math.Abs(b.Cents-want) > 1e-9 {
		t.Errorf("beat cents = %v want %v", b.Cents, want)
	}
}
