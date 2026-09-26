package meter

import (
	"math"
	"slices"
	"testing"
)

func TestAnalyzerBassMeasuresThreeOctaves(t *testing.T) {
	a := NewAnalyzer(analyzerConfig(48000, func(c *AnalyzerConfig) {
		c.Bass, c.UpperOctaves, c.Reeds = true, [2]bool{true, true}, []int{1, 2, 3}
	}))
	ring := NewSampleRing(RingCapacity)
	analyzerTone(ring, 48000, 4, 110.1, 220.3, 440.5)
	var f Frame
	for range 3 {
		f = a.Analyze(ring)
	}
	if f.Status != FrameMeasured || !slices.Equal(f.Notes, []int{33, 45, 57}) {
		t.Fatalf("frame %+v", f)
	}
	for slot, hz := range []float64{110.1, 220.3, 440.5} {
		if got := a.Display.Readouts.Frequencies[slot]; !got.Valid || math.Abs(got.Frequency-hz) > 0.02 {
			t.Fatalf("slot %d: %+v want %v", slot, got, hz)
		}
	}
	// Octave pairs beat against the nearest harmonic: 2*110.1 vs 220.3 is 0.1 Hz.
	if beat := a.Display.Readouts.Panels[1].Current; math.Abs(beat-0.1) > 0.02 {
		t.Fatalf("pair12 %v", beat)
	}
	if a.Display.Label.Text != "A" || a.Display.Label.Octave != 2 || a.Previous != 33+45+57 {
		t.Fatalf("label %+v previous %d", a.Display.Label, a.Previous)
	}
}

func TestAnalyzerChordsRecognizeAndRetain(t *testing.T) {
	a := NewAnalyzer(analyzerConfig(44100, func(c *AnalyzerConfig) {
		c.Chords, c.Reeds = true, []int{1, 2, 3}
		c.ChordLabel = func(id ChordIdentity) string {
			return "C" + map[ChordKind]string{ChordMajor: "", ChordMinor: "m"}[id.Kind]
		}
	}))
	ring := NewSampleRing(RingCapacity)
	c4, e4, g4 := 440*math.Pow(2, -9.0/12), 440*math.Pow(2, -5.0/12), 440*math.Pow(2, -2.0/12)
	analyzerTone(ring, 44100, 4, c4, e4, g4)
	var f Frame
	for range 3 {
		f = a.Analyze(ring)
	}
	if f.Status != FrameMeasured || !slices.Equal(f.Notes, []int{48, 52, 55}) || a.Display.Label != (Label{Text: "C", Octave: -1}) {
		t.Fatalf("frame %+v label %+v", f, a.Display.Label)
	}
	if a.History.LastNotes != [3]int{48, 52, 55} {
		t.Fatalf("complete chords must reach history: %+v", a.History)
	}
	for slot := range 3 {
		if !a.Display.Readouts.Frequencies[slot].Valid {
			t.Fatalf("slot %d not measured", slot)
		}
	}
}

func TestAnalyzerCompoundSilenceRequestsOneReset(t *testing.T) {
	a := NewAnalyzer(analyzerConfig(48000, func(c *AnalyzerConfig) { c.Bass, c.Reeds = true, []int{1} }))
	ring := NewSampleRing(RingCapacity)
	analyzerTone(ring, 48000, 4, 110.1)
	for range 3 {
		a.Analyze(ring)
	}
	silent := NewSampleRing(RingCapacity)
	silent.Append(make([]int16, 200000))
	resets := 0
	for range 4 {
		if a.Analyze(silent).ResetRequested {
			resets++
		}
	}
	if resets != 1 || a.Previous != 0 {
		t.Fatalf("resets %d previous %d", resets, a.Previous)
	}
}
