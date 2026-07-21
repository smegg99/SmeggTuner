package dsp

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	gaudio "github.com/go-audio/audio"
	"github.com/go-audio/wav"

	"smegg.me/smeggtuner/core/audio"
	"smegg.me/smeggtuner/core/tuning"
)

func TestEngineSeparatedReedsAndScalePitch(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 6 * time.Second,
		Reeds: []audio.ReedSpec{
			{Freq: 440.0, Amp: 0.4},
			{Freq: 442.6, Amp: 0.35},
		},
	}
	cfg := defaultCfg()
	cfg.ReedCount = 2
	m := runEngine(t, spec, cfg).lastFull()
	if !m.ReedsSeparated {
		t.Fatalf("2.6 Hz apart is well clear of the resolution limit: %+v", m.Reeds)
	}
	// DevCents is measured against this pitch, so it must be the note's exact frequency, not the band's midpoint.
	if math.Abs(m.ScalePitch-440.0) > 1e-9 {
		t.Fatalf("scale pitch = %v want 440 (A4 at A4=440)", m.ScalePitch)
	}
	want := tuning.Cents(m.Reeds[0].Freq, m.ScalePitch)
	if math.Abs(m.Reeds[0].DevCents-want) > 1e-9 {
		t.Fatalf("devCents %v disagrees with cents(freq, scalePitch) %v", m.Reeds[0].DevCents, want)
	}
}

func TestEngineManualNoteAndPPM(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 5 * time.Second,
		Reeds:    []audio.ReedSpec{{Freq: 440.0, Amp: 0.4}},
	}
	cfg := defaultCfg()
	cfg.ManualNote = tuning.NoteA4
	cfg.ClockPPM = 100 // reported = est * (1 - 1e-4) => ~439.956
	cap := runEngine(t, spec, cfg)
	m := cap.lastFull()
	if math.Abs(m.Reeds[0].Freq-439.956) > 0.05 {
		t.Fatalf("ppm not applied: %v", m.Reeds[0].Freq)
	}
}

// The engine filters block samples in place. If a FileSource hands out windows into its own decoded
// array, the first filtered pass corrupts the source and every later pass measures different audio.
func TestEngineFileSourceRepeatableUnderFilters(t *testing.T) {
	p := filepath.Join(t.TempDir(), "reuse.wav")
	writeSineWAV(t, p, 442.6, 50, 48000, 6)

	cfg := defaultCfg()
	cfg.Hum50 = true
	pass := func() Measurement {
		src, err := audio.NewFileSource(p, false, false)
		if err != nil {
			t.Fatal(err)
		}
		cap := &capture{}
		if err := NewEngine(cfg, cap.emit).Run(context.Background(), src); err != nil {
			t.Fatal(err)
		}
		// the same source object replayed: this is what a UI loop does
		cap2 := &capture{}
		if err := NewEngine(cfg, cap2.emit).Run(context.Background(), src); err != nil {
			t.Fatal(err)
		}
		first, second := cap.lastFull(), cap2.lastFull()
		if len(first.Reeds) != 1 || len(second.Reeds) != 1 {
			t.Fatalf("reeds: first=%d second=%d", len(first.Reeds), len(second.Reeds))
		}
		if first.Reeds[0].Freq != second.Reeds[0].Freq {
			t.Errorf("replay changed the measurement: %.9f then %.9f (delta %g Hz)",
				first.Reeds[0].Freq, second.Reeds[0].Freq,
				second.Reeds[0].Freq-first.Reeds[0].Freq)
		}
		if first.InputLevel != second.InputLevel {
			t.Errorf("replay changed the input level: %v then %v",
				first.InputLevel, second.InputLevel)
		}
		return first
	}
	m := pass()
	if math.Abs(m.Reeds[0].Freq-442.6) > 0.05 {
		t.Fatalf("freq = %v want 442.6 +-0.05", m.Reeds[0].Freq)
	}
}

// writeSineWAV writes a mono 16-bit sine, optionally with mains hum mixed in.
func writeSineWAV(t *testing.T, path string, freq, humFreq float64, sr, seconds int) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	enc := wav.NewEncoder(f, sr, 16, 1, 1)
	n := sr * seconds
	buf := &gaudio.IntBuffer{
		Format:         &gaudio.Format{NumChannels: 1, SampleRate: sr},
		SourceBitDepth: 16,
		Data:           make([]int, n),
	}
	for i := 0; i < n; i++ {
		tSec := float64(i) / float64(sr)
		v := 0.5 * math.Sin(2*math.Pi*freq*tSec)
		if humFreq > 0 {
			v += 0.3 * math.Sin(2*math.Pi*humFreq*tSec)
		}
		buf.Data[i] = int(30000 * v)
	}
	if err := enc.Write(buf); err != nil {
		t.Fatal(err)
	}
	if err := enc.Close(); err != nil {
		t.Fatal(err)
	}
}

// Silence must not silence the engine: the UI needs a heartbeat to tell a quiet room from a dead engine.
func TestEngineHeartbeatOnSilence(t *testing.T) {
	spec := audio.SynthSpec{Duration: 3 * time.Second} // no reeds: digital silence
	cap := runEngine(t, spec, defaultCfg())
	if len(cap.all) < 10 {
		t.Fatalf("3s of silence produced %d measurements, want >= 10", len(cap.all))
	}
	quiet := 0
	for i, m := range cap.all {
		if len(m.Reeds) != 0 {
			t.Fatalf("measurement %d reported reeds on silence: %+v", i, m.Reeds)
		}
		if m.State == StateTooQuiet {
			quiet++
		}
	}
	if quiet == 0 {
		t.Fatalf("no tooQuiet state in %d silent measurements", len(cap.all))
	}
	t.Logf("silence: %d measurements, %d tooQuiet", len(cap.all), quiet)
}

// Update and Freeze are called from UI goroutines; they must never block on a Run loop that is not draining.
func TestEngineUpdateAndFreezeDoNotBlockWithoutRun(t *testing.T) {
	e := NewEngine(defaultCfg(), func(Measurement) {})
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 20; i++ {
			e.Update(func(c *EngineConfig) { c.A4 = 442 })
			e.Freeze(i%2 == 0)
		}
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Update/Freeze blocked with no Run loop draining")
	}
}

func TestEngineInitializingState(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 1 * time.Second,
		Reeds:    []audio.ReedSpec{{Freq: 440, Amp: 0.3}},
	}
	cfg := defaultCfg()
	cfg.CalibSecs = 5 // longer than the signal: must stay initializing
	cap := runEngine(t, spec, cfg)
	for _, m := range cap.all {
		if m.State != StateInitializing {
			t.Fatalf("expected only initializing states, got %v", m.State)
		}
	}
}
