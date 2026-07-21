package tests

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/audio"
	"smegg.me/smeggtuner/core/dsp"
)

type expect struct {
	Reeds   int       `json:"reeds"`
	A4      float64   `json:"a4"`
	Note    string    `json:"note"`
	Freqs   []float64 `json:"freqs"`
	FreqTol float64   `json:"freqTol"`
	Beats   []float64 `json:"beats"`
	BeatTol float64   `json:"beatTol"`
}

func TestFixtures(t *testing.T) {
	paths, _ := filepath.Glob("fixtures/*.wav")
	if len(paths) == 0 {
		t.Skip("no WAV fixtures present yet")
	}
	for _, wavPath := range paths {
		name := strings.TrimSuffix(filepath.Base(wavPath), ".wav")
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(strings.TrimSuffix(wavPath, ".wav") + ".expect.json")
			if err != nil {
				t.Fatalf("fixture %s has no expect file: %v", name, err)
			}
			var ex expect
			if err := json.Unmarshal(raw, &ex); err != nil {
				t.Fatal(err)
			}
			if ex.A4 == 0 {
				ex.A4 = 440
			}
			if ex.FreqTol == 0 {
				ex.FreqTol = 0.1
			}
			if ex.BeatTol == 0 {
				ex.BeatTol = 0.1
			}
			src, err := audio.NewFileSource(wavPath, false, false)
			if err != nil {
				t.Fatal(err)
			}
			var all []dsp.Measurement
			eng := dsp.NewEngine(dsp.EngineConfig{
				A4: ex.A4, ReedCount: ex.Reeds,
				FineWindow: 3 * time.Second,
			}, func(m dsp.Measurement) {
				all = append(all, m)
			})
			if err := eng.Run(context.Background(), src); err != nil {
				t.Fatal(err)
			}
			// Real recordings never lock (bellows drift) and end in silence, so aggregate the whole stream.
			m, ok := dsp.Aggregate(all, ex.A4, 0.5)
			if !ok {
				t.Fatal("no usable measurements in recording")
			}
			if ex.Note != "" && m.NoteName != ex.Note {
				t.Errorf("note = %s want %s", m.NoteName, ex.Note)
			}
			if len(m.Reeds) != len(ex.Freqs) {
				t.Fatalf("reeds = %d want %d (%+v)", len(m.Reeds), len(ex.Freqs), m.Reeds)
			}
			for i, want := range ex.Freqs {
				if math.Abs(m.Reeds[i].Freq-want) > ex.FreqTol {
					t.Errorf("reed %d = %v want %v +-%v", i+1, m.Reeds[i].Freq, want, ex.FreqTol)
				}
			}
			for i, want := range ex.Beats {
				if i >= len(m.Beats) {
					t.Errorf("missing beat %d", i)
					continue
				}
				if math.Abs(m.Beats[i].Hz-want) > ex.BeatTol {
					t.Errorf("beat %d = %v want %v +-%v", i, m.Beats[i].Hz, want, ex.BeatTol)
				}
			}
		})
	}
}
