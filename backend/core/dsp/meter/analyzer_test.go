package meter

import (
	"math"
	"testing"
)

func analyzerConfig(rate int, edit func(*AnalyzerConfig)) AnalyzerConfig {
	controls := make([]NoteCurve, 3)
	for s := range controls {
		for i := range controls[s] {
			controls[s][i] = Missing
		}
	}
	cfg := AnalyzerConfig{SampleRate: rate, RateFactor: 1, Table: EqualTemperament(440, 0), Targets: BuildTargets(controls, nil),
		Reeds: []int{1}, ScaleRelative: true, Hertz: true, Octave: -1, PitchClass: -1,
		Name: func(n int) string {
			return [12]string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}[n%12]
		}}
	if edit != nil {
		edit(&cfg)
	}
	return cfg
}

func analyzerTone(ring *SampleRing, rate int, seconds float64, freqs ...float64) {
	n := int(seconds * float64(rate))
	samples := make([]int16, n)
	for i := range samples {
		v := 0.0
		for _, f := range freqs {
			v += 6000 * math.Sin(2*math.Pi*f*float64(i)/float64(rate))
		}
		samples[i] = int16(math.Round(v))
	}
	ring.Append(samples)
}

func TestAnalyzerMeasuresASteadyTone(t *testing.T) {
	a := NewAnalyzer(analyzerConfig(48000, nil))
	ring := NewSampleRing(RingCapacity)
	analyzerTone(ring, 48000, 4, 440.5)
	var f Frame
	for range 3 {
		f = a.Analyze(ring)
	}
	if f.Status != FrameMeasured || f.Note != 57 {
		t.Fatalf("frame %+v", f)
	}
	want := (440.5/440 - 1) / 0.0006
	if d := f.Refined.Deviations[0]; math.Abs(d-want) > 0.05 {
		t.Fatalf("deviation %v want %v", d, want)
	}
	if got := a.Display.Readouts.Frequencies[0]; !got.Valid || math.Abs(got.Frequency-440.5) > 0.01 {
		t.Fatalf("cache %+v", got)
	}
}

func TestAnalyzerLocksHoldsAndReleases(t *testing.T) {
	a := NewAnalyzer(analyzerConfig(44100, func(c *AnalyzerConfig) { c.StopAfterLock = true }))
	ring := NewSampleRing(RingCapacity)
	analyzerTone(ring, 44100, 4, 220.2)
	var f Frame
	for i := range 14 {
		f = a.Analyze(ring)
		if a.History.Acquired {
			break
		}
		if i == 13 {
			t.Fatalf("no acquisition: %+v %+v", f, a.History)
		}
	}
	if !a.Display.VisibleLock || !a.History.Locked {
		t.Fatal("acquisition must show the lock")
	}
	if f = a.Analyze(ring); f.Status != FrameHeld {
		t.Fatalf("stop-after-lock must hold, got %+v", f)
	}
	silent := NewSampleRing(RingCapacity)
	silent.Append(make([]int16, 200000))
	// Selection reads the previous refresh, so the first silent event still holds the locked note.
	if f = a.Analyze(silent); f.Status != FrameHeld {
		t.Fatalf("first silent event: %+v", f)
	}
	if f = a.Analyze(silent); f.Status != FrameNoNote || !f.ResetRequested || a.History.Acquired || a.Display.VisibleLock {
		t.Fatalf("silence: %+v acquired %v", f, a.History.Acquired)
	}
	if !a.Display.Readouts.Frequencies[0].Valid || !a.Display.Readouts.FrequencyDimmed {
		t.Fatal("no note keeps the number and validity but dims")
	}
	if f = a.Analyze(silent); f.ResetRequested {
		t.Fatal("a second silent frame must not request another reset")
	}
}

func TestAnalyzerSplitsThreeReeds(t *testing.T) {
	a := NewAnalyzer(analyzerConfig(48000, func(c *AnalyzerConfig) { c.Reeds = []int{1, 2, 3} }))
	ring := NewSampleRing(RingCapacity)
	analyzerTone(ring, 48000, 4, 438.9, 440.3, 441.8)
	var f Frame
	for range 3 {
		f = a.Analyze(ring)
	}
	if f.Status != FrameMeasured {
		t.Fatalf("frame %+v", f)
	}
	// Rectangular-window leakage between reeds 1.4 Hz apart biases the reference estimate by a few hundredths.
	for slot, hz := range []float64{438.9, 440.3, 441.8} {
		if got := a.Display.Readouts.Frequencies[slot].Frequency; math.Abs(got-hz) > 0.05 {
			t.Fatalf("slot %d: %v want %v", slot, got, hz)
		}
	}
	if beat := a.Display.Readouts.Panels[1].Current; math.Abs(beat-1.4) > 0.08 {
		t.Fatalf("pair12 beat %v", beat)
	}
}

func TestAnalyzerCarryKeepsTheLearnedBackground(t *testing.T) {
	ring := NewSampleRing(RingCapacity)
	ring.Append(make([]int16, 100000))
	a := NewAnalyzer(analyzerConfig(48000, nil))
	a.Learn()
	for range 9 {
		a.Analyze(ring)
	}
	b := NewAnalyzer(analyzerConfig(48000, func(c *AnalyzerConfig) { c.Reeds = []int{1, 2, 3} }))
	if !b.Carry(a) || b.coarse.Notes[50].Threshold != a.coarse.Notes[50].Threshold {
		t.Fatal("a reed-count change must keep the learned coarse state")
	}
	c := NewAnalyzer(analyzerConfig(48000, func(c *AnalyzerConfig) { c.Table = EqualTemperament(442, 0) }))
	if c.Carry(a) {
		t.Fatal("a new A4 moves every band: the background must be relearned")
	}
}
