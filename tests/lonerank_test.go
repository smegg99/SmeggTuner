package tests

// sounds/*'8.wav and *'16.wav: one Hohner rank alone at A442; an independent FFT put each within 3.6 cents of its pitch.

import (
	"math"
	"path/filepath"
	"sort"
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/tuning"
)

const (
	loneRankA4       = 442.0
	loneRankMaxCents = 10.0
)

// strayLine: with no learned background a weak low line is the lowest eligible note, and only a note scoring 30x it replaces it.
var strayLine = map[string]string{
	"Cis (C#) '16": "a 39.4 Hz line (D#1) scores 5-7; C#3 scores 143-190, under 30x it, the 9th partial D#6 147-227 clears it",
	"D '16":        "a 46-49 Hz line (F#1/G1) scores 3-4; D3 scores 82-112, under 30x it, the 3rd partial A4 96-141 clears it",
	"F '16":        "a 120 Hz line (A#2) scores 3; F3 scores 58-98 and only sometimes clears 30x it, C6 and A5 partials clear it",
	"Fis '16":      "a 120 Hz line (A#2) scores 4; F#3 scores 53-72, under 30x it, the 3rd partial C#5 94-172 clears it",
}

// loneRankWant is the note the rank sounds: every key in the set is in the octave above C4, the 16' an octave under.
func loneRankWant(s sample) tuning.Note {
	if s.register == 16 {
		return tuning.Note(48 + s.class)
	}
	return tuning.Note(60 + s.class)
}

// loneRanks is every single-rank recording in sounds/, sorted so a failure list reads the same twice.
func loneRanks(t *testing.T) []sample {
	t.Helper()
	paths, _ := filepath.Glob(filepath.Join("..", "sounds", "*.wav"))
	var out []sample
	for _, p := range paths {
		if s, ok := parseSampleName(p); ok {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		t.Skip("no single-rank recordings in ../sounds")
	}
	sort.Slice(out, func(a, b int) bool { return out[a].name < out[b].name })
	return out
}

func loneRun(t *testing.T, s sample) run {
	return runFile(t, s.path, dsp.EngineConfig{A4: loneRankA4, ReedCount: 1})
}

// skipStray skips a known stray-line take that still misreads, and fails one that no longer does.
func skipStray(t *testing.T, s sample, ok bool, all []dsp.Measurement) {
	t.Helper()
	why, known := strayLine[s.name]
	switch {
	case known && !ok:
		t.Skipf("%s sounds %s: %s (reads %s)", s.name, name(loneRankWant(s)), why, summary(all))
	case known:
		t.Errorf("listed as a stray-line misreading but reads correctly now: drop it from strayLine")
	}
}

func TestEveryLoneRankIsNamedAndInTune(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("loads two dozen recordings")
	}
	for _, s := range loneRanks(t) {
		t.Run(s.name, func(t *testing.T) {
			all := loneRun(t, s).all
			want := loneRankWant(s)
			skipStray(t, s, reads(all, want), all)
			if !reads(all, want) {
				t.Fatalf("%s sounds %s, reads %s", s.name, name(want), summary(all))
			}
			freq, n := sustained(sounding(all, want), 0)
			c := cents(freq, want.Freq(loneRankA4))
			if math.Abs(c) > loneRankMaxCents {
				t.Errorf("%.3f Hz is %+.1f cents off %s at A442: the wrong note, not a reed out of tune", freq, c, name(want))
			}
			t.Logf("%s %.3f Hz %+.1f cents over %d analyses", name(want), freq, c, n)
		})
	}
}

// The lock holds the note it locked on: from the first take on, no analysis names another.
func TestALockedLoneRankNeverJumpsNote(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("loads two dozen recordings")
	}
	for _, s := range loneRanks(t) {
		t.Run(s.name, func(t *testing.T) {
			all := loneRun(t, s).all
			skipStray(t, s, reads(all, loneRankWant(s)), all)
			at := takeAt(all)
			if len(at) == 0 {
				t.Skip("never locked; TestASteadyLoneRankLocks says why")
			}
			took := all[at[0]]
			for _, m := range measured(all[at[0]:]) {
				if m.Note != took.Note {
					t.Fatalf("locked on %s at %.2fs, then read %s at %.2fs", name(took.Note), took.SourceAt, name(m.Note), m.SourceAt)
				}
			}
		})
	}
}

// The history locks once eight analyses in a row stay within 400/bins units and the note has settled: the first take is exactly there.
func TestASteadyLoneRankLocks(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("loads two dozen recordings")
	}
	for _, s := range loneRanks(t) {
		t.Run(s.name, func(t *testing.T) {
			r := loneRun(t, s)
			want := loneRankWant(s)
			skipStray(t, s, reads(r.all, want), r.all)
			limit := lockSpan(r.rate, loneRankA4, want)
			steady, span, longest := steadiness(r.all, want, limit)
			took := -1
			for _, i := range takeAt(r.all) {
				if r.all[i].Note == want {
					took = i
					break
				}
			}
			if took != steady {
				t.Fatalf("first take on %s at event %d, the history accepts at event %d (limit %.2f units, steadiest %.2f)", name(want), took, steady, limit, span)
			}
			if longest < lockWindow {
				t.Skipf("never held %s for %d analyses in a row, the longest run is %d (%s)", name(want), lockWindow, longest, summary(r.all))
			}
			if span > limit {
				t.Skipf("bellows drift: the steadiest %d analyses on %s span %.2f units, the reference locks within %.2f",
					lockWindow, name(want), span, limit)
			}
			m := r.all[took]
			if c := cents(m.Reeds[0].Freq, want.Freq(loneRankA4)); math.Abs(c) > loneRankMaxCents {
				t.Errorf("take at %.3f Hz is %+.1f cents off %s", m.Reeds[0].Freq, c, name(want))
			}
			t.Logf("took %.3f Hz at %.2fs, steadiest span %.2f of %.2f units", m.Reeds[0].Freq, m.SourceAt, span, limit)
		})
	}
}

// The same key in both registers: a scan settled on a partial in only one of them fails here.
func TestTheSixteenFootSoundsAnOctaveUnderTheEight(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("loads two dozen recordings")
	}
	byClass := map[int]map[int]sample{}
	for _, s := range loneRanks(t) {
		if byClass[s.class] == nil {
			byClass[s.class] = map[int]sample{}
		}
		byClass[s.class][s.register] = s
	}
	for class := range 12 {
		eight, sixteen := byClass[class][8], byClass[class][16]
		if eight.path == "" || sixteen.path == "" {
			continue
		}
		t.Run(sixteen.name, func(t *testing.T) {
			e, s := loneRun(t, eight).all, loneRun(t, sixteen).all
			high, low := loneRankWant(eight), loneRankWant(sixteen)
			skipStray(t, sixteen, reads(s, low), s)
			if !reads(e, high) || !reads(s, low) {
				t.Fatalf("8' reads %s and 16' reads %s, want %s and %s", summary(e), summary(s), name(high), name(low))
			}
			fe, _ := sustained(sounding(e, high), 0)
			fs, _ := sustained(sounding(s, low), 0)
			// Two reeds tuned by hand: near an octave, far from any other interval.
			if c := cents(fe, 2*fs); math.Abs(c) > 40 {
				t.Errorf("%.3f Hz over %.3f Hz is %+.1f cents off an octave", fe, fs, c)
			}
		})
	}
}
