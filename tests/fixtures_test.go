package tests

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp"
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

func readExpect(t *testing.T, path string) expect {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("fixture has no expect file: %v", err)
	}
	ex := expect{A4: 440, FreqTol: 0.1, BeatTol: 0.1}
	if err := json.Unmarshal(raw, &ex); err != nil {
		t.Fatal(err)
	}
	return ex
}

// sustainedBeat is a pair's median beat over frames where both of its slots sounded.
func sustainedBeat(frames []dsp.Measurement, pair int) (float64, int) {
	var hz []float64
	for _, m := range frames {
		if pair < len(m.Beats) && m.Beats[pair].Present {
			hz = append(hz, m.Beats[pair].Hz)
		}
	}
	if len(hz) == 0 {
		return 0, 0
	}
	sort.Float64s(hz)
	return hz[len(hz)/2], len(hz)
}

// Each fixtures/<name>.wav is held to its expect.json; see fixtures/README.md for the reduction.
func TestFixtures(t *testing.T) {
	t.Parallel()
	paths, _ := filepath.Glob("fixtures/*.wav")
	if len(paths) == 0 {
		t.Fatal("required WAV fixtures are missing")
	}
	expectations, _ := filepath.Glob("fixtures/*.expect.json")
	for _, path := range expectations {
		wav := strings.TrimSuffix(path, ".expect.json") + ".wav"
		if _, err := os.Stat(wav); err != nil {
			t.Fatalf("fixture %s: %v", wav, err)
		}
	}
	for _, wav := range paths {
		t.Run(strings.TrimSuffix(filepath.Base(wav), ".wav"), func(t *testing.T) {
			ex := readExpect(t, strings.TrimSuffix(wav, ".wav")+".expect.json")
			all := runFile(t, wav, dsp.EngineConfig{A4: ex.A4, ReedCount: ex.Reeds}).all
			note, _ := heard(all)
			if ex.Note != "" && (name(note) != ex.Note || !reads(all, note)) {
				t.Fatalf("reads %s, want %s", summary(all), ex.Note)
			}
			frames := sounding(all, note)
			for i, want := range ex.Freqs {
				got, n := sustained(frames, i)
				if n == 0 || math.Abs(got-want) > ex.FreqTol {
					t.Errorf("slot %d at %.4f Hz over %d analyses, want %.4f +-%.2f", i+1, got, n, want, ex.FreqTol)
				}
				t.Logf("slot %d: %.4f Hz over %d sounding analyses (%+.4f)", i+1, got, n, got-want)
			}
			for i, want := range ex.Beats {
				got, n := sustainedBeat(frames, i)
				if n == 0 || math.Abs(got-want) > ex.BeatTol {
					t.Errorf("beat %d at %.4f Hz over %d analyses, want %.4f +-%.2f", i+1, got, n, want, ex.BeatTol)
				}
			}
			if ts := takes(all); len(ts) == 0 {
				t.Errorf("never locked: %s", summary(all))
			} else {
				t.Logf("take at %.2fs: %.4f Hz", ts[0].SourceAt, ts[0].Reeds[0].Freq)
			}
		})
	}
}
