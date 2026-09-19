package meter

import (
	"math"
	"testing"
)

func targetsAt(note int, values ...float64) TargetTables {
	t := TargetTables{Slots: make([]NoteCurve, len(values))}
	for s := range t.Slots {
		for i := range t.Slots[s] {
			t.Slots[s][i] = Missing
		}
		t.Slots[s][note-FirstActiveNote] = values[s]
	}
	t.Coverage[note-FirstActiveNote] = true
	return t
}

func TestAnalyzerTargetsPlaceALoneReed(t *testing.T) {
	hz := 440 * (1 + 0.0006*10) // +10 units
	for _, c := range []struct {
		targets TargetTables
		slot    int
	}{
		{targetsAt(57, -10, 0, 10), 2}, // nearest target wins
		{TargetTables{}, 0},            // without targets the reference keeps slot 1
	} {
		a := NewAnalyzer(analyzerConfig(48000, func(cfg *AnalyzerConfig) {
			cfg.Reeds, cfg.Targets, cfg.ScaleRelative = []int{1, 2, 3}, c.targets, false
		}))
		ring := NewSampleRing(RingCapacity)
		analyzerTone(ring, 48000, 4, hz)
		var f Frame
		for range 3 {
			f = a.Analyze(ring)
		}
		if f.Status != FrameMeasured || f.Refined.Bins[c.slot] == 0 {
			t.Fatalf("targets %v: bins %v", c.targets.Slots != nil, f.Refined.Bins)
		}
		if c.targets.Slots != nil {
			panel := a.Display.Readouts.Panels[2*c.slot]
			if math.Abs(panel.Current-(f.Refined.Deviations[c.slot]-10)) > 1e-9 || panel.Target != 10 {
				t.Fatalf("target-relative panel %+v deviation %v", panel, f.Refined.Deviations[c.slot])
			}
		}
	}
}
