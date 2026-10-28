// Package tests runs the reference-compatible engine end to end over recorded and synthesised audio.
package tests

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"sync"
	"testing"

	"smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// run is every event one engine run emitted, and the rate of the audio it ran over.
type run struct {
	all  []dsp.Measurement
	rate int
}

func runSource(t *testing.T, cfg dsp.EngineConfig, src audio.Source) run {
	t.Helper()
	r, err := collect(cfg, src)
	if err != nil {
		t.Fatalf("engine: %v", err)
	}
	return r
}

func collect(cfg dsp.EngineConfig, src audio.Source) (run, error) {
	r := run{rate: src.Info().SampleRate}
	// Engine normalization owns its register layout; table cases share their inputs.
	cfg.Octaves = slices.Clone(cfg.Octaves)
	eng := dsp.NewEngine(cfg, func(m dsp.Measurement) { r.all = append(r.all, m) })
	// Not a test context: memoised runs outlive the test that filled them.
	return r, eng.Run(context.Background(), src)
}

type fileRun struct {
	once sync.Once
	run  run
	err  error
}

var (
	runsMu sync.Mutex
	runs   = map[string]*fileRun{}
)

// runFile is one engine run over a recording as the app plays a file, memoised on the recording and the configuration.
func runFile(t *testing.T, path string, cfg dsp.EngineConfig) run {
	t.Helper()
	key := fmt.Sprintf("%s|%+v", path, cfg)
	runsMu.Lock()
	r := runs[key]
	if r == nil {
		r = &fileRun{}
		runs[key] = r
	}
	runsMu.Unlock()
	r.once.Do(func() {
		src, err := audio.NewFileSource(path, false, false)
		if err != nil {
			r.err = err
			return
		}
		r.run, r.err = collect(cfg, src)
	})
	if r.err != nil {
		if errors.Is(r.err, os.ErrNotExist) {
			t.Skipf("optional recording not present: %v", r.err)
		}
		t.Fatalf("recording failed: %v", r.err)
	}
	return r.run
}

// registerConfig is what the app configures for a pulled treble register named by its banks.
func registerConfig(t *testing.T, a4 float64, register string) dsp.EngineConfig {
	t.Helper()
	banks, err := session.ParseBanks(register)
	if err != nil {
		t.Fatalf("register %q: %v", register, err)
	}
	return dsp.EngineConfig{A4: a4, ReedCount: len(banks), Octaves: session.OctavesOf(banks)}
}

// bassConfig is what the app configures facing a bass machine of the given voice count.
func bassConfig(t *testing.T, a4 float64, voices int) dsp.EngineConfig {
	t.Helper()
	feet := session.BassFeet(voices)
	if feet == nil {
		t.Fatalf("%d is not a bass machine this app describes", voices)
	}
	return dsp.EngineConfig{A4: a4, ReedCount: voices, Bass: true, Octaves: session.OctavesOfFeet(feet)}
}

// takes are what the recorder files (services/record edge): the visible lock rising, or moving note, on a measured frame.
func takes(all []dsp.Measurement) []dsp.Measurement {
	var out []dsp.Measurement
	for _, i := range takeAt(all) {
		out = append(out, all[i])
	}
	return out
}

// takeAt is the index of every take in all.
func takeAt(all []dsp.Measurement) []int {
	var out []int
	locked, note := false, tuning.Note(0)
	for i, m := range all {
		if m.Frame == dsp.FrameNone {
			continue
		}
		if m.Frame == dsp.FrameMeasured && m.Lock.Visible && (!locked || m.Note != note) && len(m.Reeds) > 0 {
			out = append(out, i)
		}
		locked, note = m.Lock.Visible, m.Note
	}
	return out
}

// measured is every analysis that wrote fresh slot values, optionally only those on one note.
func measured(all []dsp.Measurement, on ...tuning.Note) []dsp.Measurement {
	var out []dsp.Measurement
	for _, m := range all {
		if m.Frame == dsp.FrameMeasured && (len(on) == 0 || m.Note == on[0]) {
			out = append(out, m)
		}
	}
	return out
}
