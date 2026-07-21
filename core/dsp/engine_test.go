package dsp

import (
	"context"
	"math"
	"sync"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/audio"
	"smegg.me/smeggtuner/core/tuning"
)

type capture struct {
	mu  sync.Mutex
	all []Measurement
}

func (c *capture) emit(m Measurement) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.all = append(c.all, m)
}

// lastFull returns the last measurement carrying a fine-stage result; the final emission is often a
// reedless heartbeat.
func (c *capture) lastFull() Measurement {
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := len(c.all) - 1; i >= 0; i-- {
		if len(c.all[i].Reeds) > 0 {
			return c.all[i]
		}
	}
	return Measurement{}
}

func runEngine(t *testing.T, spec audio.SynthSpec, cfg EngineConfig) *capture {
	t.Helper()
	cap := &capture{}
	e := NewEngine(cfg, cap.emit)
	src := audio.NewSynthSource(spec)
	if err := e.Run(context.Background(), src); err != nil {
		t.Fatal(err)
	}
	return cap
}

// sliceSource plays a fixed buffer, so a test can shape the level over time (silence, then a note).
type sliceSource struct {
	samples []float32
	rate    int
}

func (s *sliceSource) Info() audio.SourceInfo {
	return audio.SourceInfo{Name: "slice", SampleRate: s.rate}
}

func (s *sliceSource) Stop() error { return nil }

func (s *sliceSource) Start(ctx context.Context) (<-chan audio.Block, error) {
	ch := make(chan audio.Block, 4)
	go func() {
		defer close(ch)
		const block = 4096
		for off := 0; off < len(s.samples); off += block {
			end := off + block
			if end > len(s.samples) {
				end = len(s.samples)
			}
			buf := append([]float32(nil), s.samples[off:end]...)
			select {
			case <-ctx.Done():
				return
			case ch <- audio.Block{Samples: buf, SampleRate: s.rate, Time: time.Now()}:
			}
		}
	}()
	return ch, nil
}

func defaultCfg() EngineConfig {
	// CalibSecs 0: synth input starts mid-note, calibration must be off
	return EngineConfig{
		A4: 440, ReedCount: 1,
		FineWindow: 3 * time.Second, Highpass: true,
	}
}

func TestEngineSingleReed(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 6 * time.Second,
		Reeds:    []audio.ReedSpec{{Freq: 442.6, Amp: 0.4, Harmonics: []float64{0.3, 0.15}}},
	}
	cap := runEngine(t, spec, defaultCfg())
	m := cap.lastFull()
	if m.Note != tuning.NoteA4 {
		t.Fatalf("note = %v (%s)", m.Note, m.NoteName)
	}
	if len(m.Reeds) != 1 {
		t.Fatalf("reeds = %+v", m.Reeds)
	}
	if math.Abs(m.Reeds[0].Freq-442.6) > 0.05 {
		t.Fatalf("freq = %v want 442.6 +-0.05", m.Reeds[0].Freq)
	}
	wantCents := tuning.Cents(442.6, 440)
	if math.Abs(m.Reeds[0].DevCents-wantCents) > 0.2 {
		t.Fatalf("cents = %v want %v", m.Reeds[0].DevCents, wantCents)
	}
	if !m.Locked {
		t.Fatal("6s stable tone must lock")
	}
}

func TestEngineTwoReedsBeat(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 6 * time.Second,
		Reeds: []audio.ReedSpec{
			{Freq: 440.0, Amp: 0.4},
			{Freq: 442.6, Amp: 0.35},
		},
	}
	cfg := defaultCfg()
	cfg.ReedCount = 2
	cap := runEngine(t, spec, cfg)
	m := cap.lastFull()
	if len(m.Reeds) != 2 || len(m.Beats) != 1 {
		t.Fatalf("reeds=%d beats=%d", len(m.Reeds), len(m.Beats))
	}
	if math.Abs(m.Beats[0].Hz-2.6) > 0.05 {
		t.Fatalf("beat = %v want 2.6 +-0.05", m.Beats[0].Hz)
	}
	if m.Beats[0].FromEnvelope {
		t.Fatal("resolved peaks must not be FromEnvelope")
	}
}

// Reeds that beat too slowly to pull apart spectrally (1 Hz is inside the 1.33 Hz main lobe): the
// peaks are one lobe, but the envelope still hears the swing. A slower pair than this is not reported.
func TestEngineEnvelopeFallback(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 8 * time.Second,
		Reeds: []audio.ReedSpec{
			{Freq: 440.0, Amp: 0.4},
			{Freq: 441.0, Amp: 0.4},
		},
	}
	cfg := defaultCfg()
	cfg.ReedCount = 2
	cap := runEngine(t, spec, cfg)
	m := cap.lastFull()
	if len(m.Beats) != 1 || !m.Beats[0].FromEnvelope {
		t.Fatalf("expected envelope fallback, got %+v", m.Beats)
	}
	if math.Abs(m.Beats[0].Hz-1.0) > 0.06 {
		t.Fatalf("beat = %v want 1.0 +-0.06", m.Beats[0].Hz)
	}
	// The peaks here are lobes of one merged line, not two reeds.
	if m.ReedsSeparated {
		t.Fatal("merged reeds must not be reported as separated")
	}
}
