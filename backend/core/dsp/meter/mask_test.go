package meter

import (
	"math"
	"testing"
)

func maskMeasure(t *testing.T, reeds []int, analyses int, freqs ...float64) *Analyzer {
	t.Helper()
	a := NewAnalyzer(analyzerConfig(48000, func(c *AnalyzerConfig) { c.Reeds = reeds }))
	ring := takeRing(48000, freqs...)
	for range analyses {
		if f := a.Analyze(ring); f.Status != FrameMeasured && f.Status != FrameNoNote {
			t.Fatalf("frame %+v", f)
		}
	}
	return a
}

// Ascending bins go to ascending enabled boxes; a disabled box reads nothing and a pair across it no beat.
func TestReedMaskPlacesBinsOnEnabledBoxes(t *testing.T) {
	a := maskMeasure(t, []int{1, 3}, 3, 438.9, 441.8)
	f := a.Display.Readouts.Frequencies
	if !f[0].Valid || f[1].Valid || !f[2].Valid || math.Abs(f[0].Frequency-438.9) > 0.05 || math.Abs(f[2].Frequency-441.8) > 0.05 {
		t.Fatalf("boxes 1 and 3: %+v", f)
	}
	if p := a.Display.Readouts.Panels; p[1].Current != Missing || p[3].Current != Missing {
		t.Fatalf("no beat across a disabled box: %+v %+v", p[1], p[3])
	}
	a = maskMeasure(t, []int{2}, 3, 440.5)
	if f := a.Display.Readouts.Frequencies; f[0].Valid || !f[1].Valid || math.Abs(f[1].Frequency-440.5) > 0.02 {
		t.Fatalf("box 2 alone: %+v", f)
	}
}

// With box 1 off, slot one is the -99 sentinel, so the lock comes after eight analyses whatever the others do.
func TestReedMaskWithoutBoxOneLocksOnTheSentinel(t *testing.T) {
	a := maskMeasure(t, []int{2, 3}, 12, 439.0, 441.5)
	if !a.History.Locked || a.History.AcceptedNotes != [3]int{57, 0, 0} || a.History.Values[0][0] != Missing {
		t.Fatalf("history %+v", a.History)
	}
	if beat := a.Display.Readouts.Panels[3].Current; math.Abs(beat-2.5) > 0.1 {
		t.Fatalf("pair 2-3 beat %v", beat)
	}
}

// The reference Octaves layout with box 2 off measures base and base+24 but never completes the history.
func TestBassSkippedOctaveMeasuresButNeverLocks(t *testing.T) {
	a := NewAnalyzer(analyzerConfig(48000, func(c *AnalyzerConfig) { c.Bands = []Band{{0, 1}, {24, 1}} }))
	if !a.cfg.Bass || a.cfg.UpperOctaves != [2]bool{false, true} || len(a.cfg.Reeds) != 2 {
		t.Fatalf("layout %+v", a.cfg)
	}
	ring := takeRing(48000, 110.1, 440.5)
	var f Frame
	for range 12 {
		f = a.Analyze(ring)
	}
	if f.Status != FrameMeasured || f.Notes[0] != 33 || f.Notes[1] != 0 || f.Notes[2] != 57 || a.History.Locked {
		t.Fatalf("frame %+v lock %v", f, a.History.Locked)
	}
	r := a.Display.Readouts
	if !r.Frequencies[0].Valid || r.Frequencies[1].Valid || math.Abs(r.Frequencies[2].Frequency-440.5) > 0.02 || r.Panels[3].Current != Missing {
		t.Fatalf("readouts %+v", r.Frequencies)
	}
}
