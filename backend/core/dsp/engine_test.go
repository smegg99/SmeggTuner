package dsp

import (
	"context"
	"math"
	"testing"
	"time"

	"smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/tuning"
)

func runSynth(t *testing.T, cfg EngineConfig, seconds float64, reeds ...float64) []Measurement {
	t.Helper()
	spec := audio.SynthSpec{SampleRate: 48000, Duration: time.Duration(seconds * float64(time.Second))}
	for _, f := range reeds {
		spec.Reeds = append(spec.Reeds, audio.ReedSpec{Freq: f, Amp: 0.2})
	}
	var out []Measurement
	e := NewEngine(cfg, func(m Measurement) { out = append(out, m) })
	if err := e.Run(context.Background(), audio.NewSynthSource(spec)); err != nil {
		t.Fatal(err)
	}
	return out
}

func lastMeasured(t *testing.T, ms []Measurement) Measurement {
	t.Helper()
	for i := len(ms) - 1; i >= 0; i-- {
		if ms[i].Frame == FrameMeasured {
			return ms[i]
		}
	}
	t.Fatal("no measured frame")
	return Measurement{}
}

func TestEngineMeasuresOneReed(t *testing.T) {
	ms := runSynth(t, EngineConfig{}, 6, 440.5)
	m := lastMeasured(t, ms)
	if m.Note != tuning.Note(69) || m.ScalePitch != 440 || len(m.Reeds) != 1 || len(m.Panels) != 1 || len(m.Needles) != 1 {
		t.Fatalf("shape %+v", m)
	}
	want := (440.5/440 - 1) / 0.0006
	if r := m.Reeds[0]; !r.Present || math.Abs(r.Dev-want) > 0.05 || math.Abs(r.Freq-440.5) > 0.01 {
		t.Fatalf("reed %+v", r)
	}
	if !m.Needles[0].Valid || m.State != StateRunning || len(m.Spectrum) != SpectrumColumns {
		t.Fatalf("needle %+v state %s", m.Needles[0], m.State)
	}
	// Heartbeats fill the blocks no analysis took; either way the screen hears every block or so.
	if len(ms) < 50 {
		t.Fatalf("%d updates over six seconds", len(ms))
	}
}

func TestEngineThreeReedsAndBeats(t *testing.T) {
	m := lastMeasured(t, runSynth(t, EngineConfig{ReedCount: 3, BeatsInHertz: true}, 6, 438.9, 440.3, 441.8))
	if len(m.Reeds) != 3 || len(m.Beats) != 2 || len(m.Panels) != 5 {
		t.Fatalf("shape reeds %d beats %d panels %d", len(m.Reeds), len(m.Beats), len(m.Panels))
	}
	for i, hz := range []float64{438.9, 440.3, 441.8} {
		if math.Abs(m.Reeds[i].Freq-hz) > 0.05 {
			t.Fatalf("reed %d %+v", i, m.Reeds[i])
		}
	}
	if b := m.Beats[0]; !b.Present || math.Abs(b.Hz-1.4) > 0.08 || math.Abs(m.Panels[1].Value-b.Hz) > 1e-9 {
		t.Fatalf("beat %+v panel %+v", b, m.Panels[1])
	}
}

// scriptSource plays silence, then a tone, as one continuous capture.
type scriptSource struct{ silence, tone float64 }

func (s scriptSource) Info() audio.SourceInfo {
	return audio.SourceInfo{Name: "script", SampleRate: 48000}
}
func (s scriptSource) Stop() error { return nil }
func (s scriptSource) Start(ctx context.Context) (<-chan audio.Block, error) {
	ch := make(chan audio.Block)
	go func() {
		defer close(ch)
		total, quiet := int((s.silence+s.tone)*48000), int(s.silence*48000)
		for at := 0; at < total; at += 4096 {
			b := make([]float32, min(4096, total-at))
			for i := range b {
				if at+i >= quiet {
					b[i] = float32(0.2 * math.Sin(2*math.Pi*220.2*float64(at+i)/48000))
				}
			}
			select {
			case ch <- audio.Block{Samples: b, SampleRate: 48000}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

func TestEngineLearnsOnStartAndLocks(t *testing.T) {
	var ms []Measurement
	e := NewEngine(EngineConfig{LearnOnStart: true}, func(m Measurement) { ms = append(ms, m) })
	if err := e.Run(context.Background(), scriptSource{silence: 5, tone: 8}); err != nil {
		t.Fatal(err)
	}
	if ms[0].State != StateInitializing || !ms[0].Learning {
		t.Fatalf("first frame %+v", ms[0])
	}
	m := lastMeasured(t, ms)
	if m.Learning || !m.Lock.Visible || !m.Lock.Acquired || m.Note != tuning.Note(57) {
		t.Fatalf("after learning: %+v", m.Lock)
	}
}

func TestEngineBassFeetUseTheReferenceBass(t *testing.T) {
	cfg := EngineConfig{Bass: true, Octaves: []OctaveRequest{{0, 1}, {12, 1}, {24, 1}}}
	m := lastMeasured(t, runSynth(t, cfg, 6, 110.1, 220.3, 440.5))
	if m.Note != tuning.Note(45) || len(m.Reeds) != 3 {
		t.Fatalf("bass %+v", m)
	}
	for i, oct := range []int{0, 12, 24} {
		if !m.Reeds[i].Present || m.Reeds[i].Octave != oct {
			t.Fatalf("foot %d %+v", i, m.Reeds[i])
		}
	}
}

func TestEngineTrebleRegisterNamesTheKey(t *testing.T) {
	cfg := EngineConfig{Octaves: []OctaveRequest{{-12, 1}, {0, 2}, {12, 1}}}
	m := lastMeasured(t, runSynth(t, cfg, 6, 220.1, 439.5, 441.5, 880.4))
	if m.Note != tuning.Note(69) || len(m.Reeds) != 4 {
		t.Fatalf("register %+v", m)
	}
	if m.Reeds[0].Octave != -12 || m.Reeds[1].Octave != 0 || m.Reeds[3].Octave != 12 {
		t.Fatalf("octaves %+v", m.Reeds)
	}
}

// A silent 4' still sits an octave above the key, so a hand edit can place it.
func TestEngineSilentSlotKeepsItsOctave(t *testing.T) {
	cfg := EngineConfig{Octaves: []OctaveRequest{{-12, 1}, {0, 1}, {12, 1}}}
	m := lastMeasured(t, runSynth(t, cfg, 6, 220.1, 439.5))
	if len(m.Reeds) != 3 || !m.Reeds[1].Present || m.Reeds[2].Present {
		t.Fatalf("reeds %+v, want the 4' silent", m.Reeds)
	}
	if m.Reeds[2].Octave != 12 || m.Reeds[2].Dev != meter.Missing {
		t.Fatalf("silent 4' = %+v, want octave 12 and no deviation", m.Reeds[2])
	}
}

func TestEngineFreezeHoldsTheLastResult(t *testing.T) {
	var out []Measurement
	e := NewEngine(EngineConfig{}, func(m Measurement) { out = append(out, m) })
	e.Freeze(true)
	spec := audio.SynthSpec{SampleRate: 48000, Duration: 6 * time.Second, Reeds: []audio.ReedSpec{{Freq: 440.5, Amp: 0.2}}}
	if err := e.Run(context.Background(), audio.NewSynthSource(spec)); err != nil {
		t.Fatal(err)
	}
	// Freeze sent before Run is drained by Run itself, then holds from the first measured frame on.
	frozen := 0
	for _, m := range out {
		if m.State == StateFrozen {
			frozen++
		}
	}
	if frozen == 0 {
		t.Fatal("freeze never held a result")
	}
}

func TestShownMatchesTheReferencePrequantization(t *testing.T) {
	cases := []struct {
		v      float64
		d      int
		signed bool
		want   string
	}{
		{0.15, 1, true, "+0.2"}, {-0.15, 1, true, "-0.2"}, {0.04, 1, true, "0.0"}, {-99, 1, true, "-"},
		{-0.5, 2, false, "-"}, {-0.04, 1, true, "0.0"}, {1.005, 2, false, "1.00"}, {2.345, 2, false, "2.35"}, {12.25, 1, true, "+12.3"},
	}
	for _, c := range cases {
		if got := shown(c.v, c.d, c.signed); got != c.want {
			t.Fatalf("shown(%v, %d, %v) = %q want %q", c.v, c.d, c.signed, got, c.want)
		}
	}
}

func TestTargetsFollowTheKeyIntoTheLowestBand(t *testing.T) {
	var curve meter.NoteCurve
	for i := range curve {
		curve[i] = float64(i)
	}
	cfg := EngineConfig{Octaves: []OctaveRequest{{-12, 1}, {0, 1}}, Targets: []meter.NoteCurve{curve, curve}}
	cfg.TargetCoverage[57-4] = true
	got := cfg.targets()
	// Key A4 (index 57) analyses on A3 (index 45): A3's row must hold A4's targets.
	if got.Slots[0][45-4] != float64(57-4) || !got.Coverage[45-4] || got.Coverage[57-4] {
		t.Fatalf("shifted targets %v %v", got.Slots[0][45-4], got.Coverage[45-4])
	}
}
