package meter

import (
	"math"
	"testing"
)

func registerMeasure(t *testing.T, cfg AnalyzerConfig, tones ...float64) *Analyzer {
	t.Helper()
	a := NewAnalyzer(cfg)
	ring := NewSampleRing(RingCapacity)
	analyzerTone(ring, cfg.SampleRate, 4, tones...)
	var f Frame
	for range 3 {
		f = a.Analyze(ring)
	}
	if f.Status != FrameMeasured {
		t.Fatalf("frame %+v", f)
	}
	return a
}

func registerExpect(t *testing.T, a *Analyzer, tolerance float64, hz ...float64) {
	t.Helper()
	for slot, want := range hz {
		if got := a.Display.Readouts.Frequencies[slot]; !got.Valid || math.Abs(got.Frequency-want) > tolerance {
			t.Fatalf("slot %d: %+v want %v", slot, got, want)
		}
	}
}

func TestAnalyzerFourReedExtension(t *testing.T) {
	cfg := analyzerConfig(48000, func(c *AnalyzerConfig) { c.Reeds = []int{1, 2, 3, 4} })
	a := registerMeasure(t, cfg, 437.0, 439.0, 441.0, 443.0)
	if len(a.Display.Readouts.Frequencies) != 4 || len(a.Display.Readouts.Panels) != 7 {
		t.Fatalf("four slots need seven panels: %d %d", len(a.Display.Readouts.Frequencies), len(a.Display.Readouts.Panels))
	}
	registerExpect(t, a, 0.05, 437.0, 439.0, 441.0, 443.0)
}

func TestAnalyzerOctaveLayoutOfSingleReedsIsDirkBass(t *testing.T) {
	cfg := analyzerConfig(48000, func(c *AnalyzerConfig) { c.Bands = []Band{{0, 1}, {12, 1}, {24, 1}} })
	a := registerMeasure(t, cfg, 110.1, 220.3, 440.5)
	if !a.cfg.Bass || a.cfg.UpperOctaves != [2]bool{true, true} || len(a.cfg.Bands) != 0 {
		t.Fatalf("layout must run the shared Bass path: %+v", a.cfg)
	}
	registerExpect(t, a, 0.02, 110.1, 220.3, 440.5)
	if _, ok := DirkBass([]Band{{0, 1}, {0, 1}}); ok {
		t.Fatal("two bands at one offset are not Bass")
	}
	if upper, ok := DirkBass([]Band{{0, 1}, {24, 1}}); !ok || upper != [2]bool{false, true} {
		t.Fatal("base plus two octaves with box 2 off is Bass")
	}
	if upper, ok := DirkBass([]Band{{0, 1}, {12, 1}}); !ok || upper != [2]bool{true, false} {
		t.Fatal("base plus one octave is Bass")
	}
	for _, bands := range [][]Band{{{0, 1}, {36, 1}}, {{0, 1}, {24, 1}, {12, 1}}, {{0, 1}, {12, 2}}, {{0, 2}, {12, 1}}} {
		if _, ok := DirkBass(bands); ok {
			t.Fatalf("%v is the extension", bands)
		}
	}
}

func TestAnalyzerRegisterExtensionSplitsAMusetteBand(t *testing.T) {
	cfg := analyzerConfig(48000, func(c *AnalyzerConfig) { c.Bands = []Band{{0, 1}, {12, 2}, {24, 1}} })
	a := registerMeasure(t, cfg, 110.1, 219.5, 221.5, 440.4)
	if a.cfg.Bass || len(a.Display.Readouts.Frequencies) != 4 {
		t.Fatalf("extension expected: bass %v slots %d", a.cfg.Bass, len(a.Display.Readouts.Frequencies))
	}
	registerExpect(t, a, 0.05, 110.1, 219.5, 221.5, 440.4)
	if a.History.LastNotes != [3]int{33, 45, 57} {
		t.Fatalf("history tuple %v", a.History.LastNotes)
	}
	if beat := a.Display.Readouts.Panels[3].Current; math.Abs(beat-2.0) > 0.1 {
		t.Fatalf("musette beat inside the band %v", beat)
	}
}

func TestAssignPeaksNKeepsOrderNearTargets(t *testing.T) {
	// Two peaks on four slots land on the slots whose targets they are nearest, in order.
	targets := []float64{-10, 0, 10, 20}
	got := AssignPeaksN([]int{0, 212, 0, 0}, targets, 200)
	if got[3] != 212 && got[2] != 212 {
		t.Fatalf("single high peak: %v", got)
	}
	// -8.3 and +8.3 units sit nearest the -10 and +10 targets.
	got = AssignPeaksN([]int{199, 201, 0, 0}, targets, 200)
	if got[0] != 199 || got[2] != 201 || got[1] != 0 {
		t.Fatalf("two peaks: %v", got)
	}
}
