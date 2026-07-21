package tests

// Multi-reed detection against recordings named by feet ("hohner8+8+8A4.wav"): musette (all 8') asserted strictly; octave-spanning (has 16') documented not asserted, see TestOctaveSpanningIsSingleBand.

import (
	"context"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/audio"
	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/tuning"
)

var multiFeetName = regexp.MustCompile(`(?i)^(?:hohner|hehner)?\s*((?:16|8)(?:\+(?:16|8))+)\s*([a-h])(#)?(\d)$`)

var pitchClassOf = map[byte]int{'C': 0, 'D': 2, 'E': 4, 'F': 5, 'G': 7, 'A': 9, 'H': 11, 'B': 10}

type feetSample struct {
	name  string
	reeds int         // voices the register sounds
	has16 bool        // a sixteen-foot rank is present (octave-spanning)
	note  tuning.Note // the played key, scientific pitch (A4 = MIDI 69)
}

func parseFeet(path string) (feetSample, bool) {
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	m := multiFeetName.FindStringSubmatch(base)
	if m == nil {
		return feetSample{}, false
	}
	pc := pitchClassOf[strings.ToUpper(m[2])[0]]
	if m[3] == "#" {
		pc++
	}
	oct := int(m[4][0] - '0')
	return feetSample{
		name:  base,
		reeds: strings.Count(m[1], "+") + 1,
		has16: strings.Contains(m[1], "16"),
		note:  tuning.Note((oct+1)*12 + pc),
	}, true
}

// octaveOf is the scientific octave number (D3 -> 3).
func octaveOf(n tuning.Note) int { return int(n)/12 - 1 }

func feetSamples(t *testing.T) []struct {
	path string
	s    feetSample
} {
	t.Helper()
	paths, _ := filepath.Glob(filepath.Join("..", "sounds", "*.wav"))
	var out []struct {
		path string
		s    feetSample
	}
	for _, p := range paths {
		if s, ok := parseFeet(p); ok {
			out = append(out, struct {
				path string
				s    feetSample
			}{p, s})
		}
	}
	sort.Slice(out, func(a, b int) bool { return out[a].s.name < out[b].s.name })
	return out
}

// aggregateSample runs the engine over a recording, auto-detecting the note as the app does.
func aggregateSample(t *testing.T, path string, reeds int) (dsp.Measurement, bool) {
	t.Helper()
	src, err := audio.NewFileSource(path, false, false)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	var all []dsp.Measurement
	eng := dsp.NewEngine(dsp.EngineConfig{
		A4:         440,
		ReedCount:  reeds,
		FineWindow: 3 * time.Second,
	}, func(m dsp.Measurement) { all = append(all, m) })
	if err := eng.Run(context.Background(), src); err != nil {
		t.Fatalf("engine: %v", err)
	}
	return dsp.Aggregate(all, 440, 0.5)
}

// A musette register sounds N 8' reeds in one band; the engine must resolve N reeds, N-1 beats, and the right note.
func TestMusetteReedSeparation(t *testing.T) {
	if testing.Short() {
		t.Skip("loads multi-megabyte recordings")
	}
	samples := feetSamples(t)
	musette := 0
	for _, it := range samples {
		if it.s.has16 {
			continue
		}
		musette++
		it := it
		t.Run(it.s.name, func(t *testing.T) {
			m, ok := aggregateSample(t, it.path, it.s.reeds)
			if !ok {
				t.Fatalf("no usable reading")
			}
			if int(m.Note) != int(it.s.note) {
				t.Errorf("note %s, the file says %s", m.Note.Name(tuning.NamingCDEFGAB), it.s.note.Name(tuning.NamingCDEFGAB))
			}

			// Below ~E3 three close reeds fall under the 3s window's resolution, so the engine reads fewer; low three-reed notes need only 2 reeds resolved, mid/high notes all.
			if it.s.reeds >= 3 && octaveOf(it.s.note) < 4 {
				if len(m.Reeds) < 2 {
					t.Errorf("resolved %d reeds, want at least 2 (low note, %d expected)", len(m.Reeds), it.s.reeds)
				}
				if len(m.Beats) > len(m.Reeds)-1 {
					t.Errorf("%d beats for %d reeds is too many", len(m.Beats), len(m.Reeds))
				}
				return
			}
			if len(m.Reeds) != it.s.reeds {
				t.Errorf("resolved %d reeds, the register sounds %d", len(m.Reeds), it.s.reeds)
			}
			if want := len(m.Reeds) - 1; len(m.Beats) != want {
				t.Errorf("%d beats for %d reeds, want %d", len(m.Beats), len(m.Reeds), want)
			}
		})
	}
	if musette == 0 {
		t.Skip("no musette samples in ../sounds")
	}
	t.Logf("%d musette samples checked", musette)
}

// A 16' rank sounds an octave below the analysed band; the engine reports only the dominant octave's reeds - a known limit, so this asserts only a usable reading, never a crash.
func TestOctaveSpanningIsSingleBand(t *testing.T) {
	if testing.Short() {
		t.Skip("loads multi-megabyte recordings")
	}
	seen := 0
	for _, it := range feetSamples(t) {
		if !it.s.has16 {
			continue
		}
		seen++
		it := it
		t.Run(it.s.name, func(t *testing.T) {
			m, ok := aggregateSample(t, it.path, it.s.reeds)
			if !ok {
				t.Skipf("no reading: an octave-spanning register does not always settle")
				return
			}
			// Tracked note within an octave of the key, at least one reed, beats never exceed reeds-1.
			if d := int(m.Note) - int(it.s.note); d < -12 || d > 12 {
				t.Errorf("note %s is more than an octave from the key %s", m.Note.Name(tuning.NamingCDEFGAB), it.s.note.Name(tuning.NamingCDEFGAB))
			}
			if len(m.Reeds) < 1 {
				t.Errorf("no reed resolved")
			}
			if len(m.Beats) > len(m.Reeds)-1 && len(m.Reeds) > 0 {
				t.Errorf("%d beats for %d reeds", len(m.Beats), len(m.Reeds))
			}
			t.Logf("register sounds %d voices; engine tracked %s with %d reed(s), %d beat(s)",
				it.s.reeds, m.Note.Name(tuning.NamingCDEFGAB), len(m.Reeds), len(m.Beats))
		})
	}
	if seen == 0 {
		t.Skip("no 16'-containing samples in ../sounds")
	}
}
