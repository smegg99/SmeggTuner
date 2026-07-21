package dsp

import (
	"math"
	"testing"

	"smegg.me/smeggtuner/core/tuning"
)

func TestAggregateLocked(t *testing.T) {
	ms := []Measurement{
		aggTick(tuning.NoteA4, 0.3, 441.5),
		aggTick(tuning.NoteA4, 0.3, 441.5),
	}
	if m, _ := Aggregate(ms, 440, 0.5); m.Locked {
		t.Fatal("no source tick locked")
	}
	ms[0].Locked = true
	if m, _ := Aggregate(ms, 440, 0.5); !m.Locked {
		t.Fatal("locked must be true when a tick in the sustained region locked")
	}
}

// Lock must describe the sustained region, not the recording's noise tail.
func TestAggregateLockedIgnoresTicksOutsideRegion(t *testing.T) {
	ms := []Measurement{
		aggTick(tuning.NoteA4, 0.3, 441.5),
		aggTick(tuning.NoteA4, 0.3, 441.6),
	}
	// noise tail: below the level threshold, different note, and locked
	tail := aggTick(tuning.Note(35), 0.004, 118.7)
	tail.Locked = true
	// dropout tick inside the region but with a minority reed count
	dropout := aggTick(tuning.NoteA4, 0.3, 441.5, 444.0)
	dropout.Locked = true
	ms = append(ms, tail, tail, dropout)

	m, ok := Aggregate(ms, 440, 0.5)
	if !ok {
		t.Fatal("expected ok")
	}
	if m.Note != tuning.NoteA4 {
		t.Fatalf("note = %s want A4", m.NoteName)
	}
	if m.Locked {
		t.Fatal("a lock outside the selected sustained region must be ignored")
	}
}

// The fields a display leans on must survive aggregation: ScalePitch (the reference line) and
// ReedsSeparated (whether reeds may be read one at a time).
func TestAggregateKeepsDisplayFields(t *testing.T) {
	ticks := []Measurement{
		{
			State: StateRunning, InputLevel: 0.4, Note: 69, NoteName: "A4",
			ScalePitch: 442.0, ReedsSeparated: true,
			Reeds:    []ReedMeasure{{Freq: 442.5, DevCents: 2.0}},
			Spectrum: []float32{1},
		},
		{
			State: StateRunning, InputLevel: 0.4, Note: 69, NoteName: "A4",
			ScalePitch: 442.0, ReedsSeparated: true,
			Reeds:    []ReedMeasure{{Freq: 442.7, DevCents: 2.8}},
			Spectrum: []float32{1},
		},
	}

	out, ok := Aggregate(ticks, 442.0, 0.5)
	if !ok {
		t.Fatal("no aggregate")
	}
	if out.ScalePitch != 442.0 {
		t.Errorf("scale pitch = %v, want 442: the spectrum would draw its reference at zero", out.ScalePitch)
	}
	if !out.ReedsSeparated {
		t.Error("reeds were separated, but the aggregate says otherwise")
	}
}

// A recording carries more PICTURES than readings (the spectrum ships before a reed can be measured),
// and those note-only ticks must not vote on how many reeds the note has: letting them vote elected
// "no reeds" and every real recording aggregated to nothing.
func TestAggregateIgnoresReedlessTicks(t *testing.T) {
	pictures := 8
	readings := 3

	var ticks []Measurement
	for range pictures {
		ticks = append(ticks, Measurement{
			Note:       tuning.NoteA4,
			NoteName:   "A4",
			State:      StateRunning,
			InputLevel: 0.4,
			ScalePitch: 440,
			Spectrum:   make([]float32, SpectrumColumns),
		})
	}
	for range readings {
		ticks = append(ticks, aggTick(tuning.NoteA4, 0.4, 441.2))
	}

	out, ok := Aggregate(ticks, 440, 0.5)
	if !ok {
		t.Fatal("a recording with readings in it must aggregate, however many pictures outnumber them")
	}
	if len(out.Reeds) != 1 {
		t.Fatalf("reeds = %d, want 1: the reedless ticks outvoted the readings", len(out.Reeds))
	}
	if math.Abs(out.Reeds[0].Freq-441.2) > 0.01 {
		t.Fatalf("freq = %v, want 441.2", out.Reeds[0].Freq)
	}
}

// A recording that never measured a reed aggregates to nothing, not to a confident measurement of no reeds.
func TestAggregateWithoutAnyReedFails(t *testing.T) {
	ticks := []Measurement{
		{Note: tuning.NoteA4, State: StateRunning, InputLevel: 0.4, ScalePitch: 440},
		{Note: tuning.NoteA4, State: StateRunning, InputLevel: 0.4, ScalePitch: 440},
	}
	if _, ok := Aggregate(ticks, 440, 0.5); ok {
		t.Fatal("ticks that never reached a reed must not aggregate into one")
	}
}
