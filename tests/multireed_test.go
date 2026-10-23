package tests

// Registers named by their feet and key ("hohner16+8A4.wav", A440); the name is the truth, so key and voices are checked.

import (
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/tuning"
)

var multiFeetName = regexp.MustCompile(`(?i)^(?:hohner|hehner)?\s*((?:16|8)(?:\+(?:16|8))+)\s*([a-h])(#)?(\d)$`)

var pitchClassOf = map[byte]int{'C': 0, 'D': 2, 'E': 4, 'F': 5, 'G': 7, 'A': 9, 'H': 11, 'B': 10}

// namedBySixteen are named by the 16' pitch: hohner16+8+8+8D3's 8' ranks measure D4, like the D4 take.
var namedBySixteen = map[string]bool{"hohner16+8+8+8D3": true}

type feetSample struct {
	path, name string
	reeds      int         // voices the register sounds
	has16      bool        // a sixteen-foot rank is present
	key        tuning.Note // the played key
}

func feetSamples(t *testing.T, sixteen bool) []feetSample {
	t.Helper()
	paths, _ := filepath.Glob(filepath.Join("..", "sounds", "*.wav"))
	var out []feetSample
	for _, p := range paths {
		base := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
		m := multiFeetName.FindStringSubmatch(base)
		if m == nil || strings.Contains(m[1], "16") != sixteen {
			continue
		}
		pc := pitchClassOf[strings.ToUpper(m[2])[0]]
		if m[3] == "#" {
			pc++
		}
		key := tuning.Note((int(m[4][0]-'0')+1)*12 + pc)
		if namedBySixteen[base] {
			key += 12
		}
		out = append(out, feetSample{path: p, name: base, reeds: strings.Count(m[1], "+") + 1, has16: sixteen, key: key})
	}
	if len(out) == 0 {
		t.Skip("no matching recordings in ../sounds")
	}
	sort.Slice(out, func(a, b int) bool { return out[a].name < out[b].name })
	return out
}

// config is what the app imposes: one band of N reeds for a musette, the banks' layout once a 16' sounds.
func (s feetSample) config(t *testing.T) dsp.EngineConfig {
	if s.has16 {
		return registerConfig(t, 440, "L"+strings.Repeat("M", s.reeds-1))
	}
	return dsp.EngineConfig{A4: 440, ReedCount: s.reeds}
}

// The filed take: the key, every declared voice in its slot and octave, and a beat for every neighbouring pair.
func assertRegisterTake(t *testing.T, s feetSample) {
	t.Helper()
	all := runFile(t, s.path, s.config(t)).all
	if !reads(all, s.key) {
		t.Fatalf("reads %s, the key is %s", summary(all), name(s.key))
	}
	ts := takes(all)
	if len(ts) == 0 {
		t.Fatalf("never locked: %s", summary(all))
	}
	m := ts[0]
	if m.Note != s.key || len(m.Reeds) != s.reeds || len(m.Beats) != s.reeds-1 {
		t.Fatalf("took %s with %d slots and %d beats, want %s with %d", name(m.Note), len(m.Reeds), len(m.Beats), name(s.key), s.reeds)
	}
	for i, r := range m.Reeds {
		octave := 0
		if s.has16 && i == 0 {
			octave = -12
		}
		if !r.Present || r.Octave != octave {
			t.Errorf("slot %d %+v, want a voice at octave %d", i+1, r, octave)
		}
	}
	for _, b := range m.Beats {
		if !b.Present {
			t.Errorf("beat %s missing", b.Pair)
		}
	}
	t.Logf("take at %.2fs:%s", m.SourceAt, slotsText(m))
}

func TestMusetteReedSeparation(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("loads recordings")
	}
	for _, s := range feetSamples(t, false) {
		t.Run(s.name, func(t *testing.T) { assertRegisterTake(t, s) })
	}
}

func TestOctaveSpanningRegisters(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("loads recordings")
	}
	for _, s := range feetSamples(t, true) {
		t.Run(s.name, func(t *testing.T) { assertRegisterTake(t, s) })
	}
}
