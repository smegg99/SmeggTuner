package tests

// The Bugari set: takes from a professional double-cassotto instrument sitting at A442, in
// sounds/bugari. They carry what the older Hohner recordings do not - a 16' rank whose fundamental
// is 24 dB under its loudest partial, a 4' rank tuned onto a 16's double octave, and a four-voice
// LMMH register. File names are the played key in Polish octave notation (c1 = middle C, so C1 is
// C4 and C3 is C6) followed by the register's feet; kanal is the tone chamber.

import (
	"context"
	"math"
	"path/filepath"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/audio"
	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/tuning"
)

// The instrument's own reference. Judging it against A440 would only report it sharp.
const bugariA4 = 442.0

// Frequencies are measured to a hundredth of a hertz across a run; the tolerance is the drift a
// worked bellows puts on top of that.
const (
	bugariFreqTol = 0.05
	bugariBeatTol = 0.05
)

type bugariRank struct {
	freq   float64
	octave int // semitones from the key: -12 a 16', 0 an 8', +12 a 4'
}

type bugariBand struct {
	octave, ranks, found int
	ghostOnly            bool
}

type bugariBeat struct {
	pair string
	hz   float64
}

type bugariTake struct {
	file     string
	register string // the banks the register sounds, as the card names them
	key      string
	reeds    []bugariRank
	bands    []bugariBand
	beats    []bugariBeat
	// why, when the take does not resolve every declared rank
	short string
}

var bugariTakes = []bugariTake{{
	file: "B1 16.wav", register: "L", key: "A#4",
	reeds: []bugariRank{{234.145, -12}},
	bands: []bugariBand{{-12, 1, 1, false}},
}, {
	file: "C1 16.wav", register: "L", key: "C4",
	reeds: []bugariRank{{131.286, -12}},
	bands: []bugariBand{{-12, 1, 1, false}},
}, {
	// One rank in the key's own octave: no compound layout, the single band answers.
	file: "C1 8kanal.wav", register: "M1", key: "C4",
	reeds: []bugariRank{{262.799, 0}},
}, {
	// The 4' sounds 1.41 Hz under the 8's second partial, so the octave beat is signed negative.
	file: "C3 4 8.wav", register: "MH", key: "C6",
	reeds: []bugariRank{{1051.356, 0}, {2101.307, 12}},
	bands: []bugariBand{{0, 1, 1, false}, {12, 1, 1, false}},
	beats: []bugariBeat{{"1-2", -1.408}},
}, {
	// The 4' rank is sounding - the take's partials at k=4, 12 and 16 sit off the 16's exact
	// multiples where every other partial lands on them - but it is only 0.5 cents over the
	// double octave, a 0.17 Hz beat. That is a quarter of minBeatHz, so no window this engine
	// uses can separate the two, and the band reports the rank as harmonic-only rather than
	// inventing a reading for it.
	file: "D1   4 16.wav", register: "LH", key: "D4",
	reeds: []bugariRank{{147.377, -12}},
	bands: []bugariBand{{-12, 1, 1, false}, {12, 1, 0, true}},
	short: "the 4' rank is within half a cent of the 16's double octave",
}, {
	// Four declared ranks, three lines: the two 8' ranks are entrained, one carrier with sidebands
	// at a fixed 1.15 Hz that do not scale with harmonic number. The beat reported is the 8'-to-4'
	// octave beat, not a musette one.
	file: "G1   4 88 16.wav", register: "LMMH", key: "G4",
	reeds: []bugariRank{{197.093, -12}, {393.399, 0}, {787.972, 12}},
	bands: []bugariBand{{-12, 1, 1, false}, {0, 2, 1, false}, {12, 1, 1, false}},
	beats: []bugariBeat{{"2-3", 1.168}},
	short: "the two 8' ranks sound entrained, as one line",
}}

func bugariPath(file string) string { return filepath.Join("..", "sounds", "bugari", file) }

// measureBugari runs a take through the engine with the register its name declares, exactly as the
// app would from the card's banks.
func measureBugari(t *testing.T, take bugariTake) (dsp.Measurement, bool) {
	t.Helper()
	banks, err := session.ParseBanks(take.register)
	if err != nil {
		t.Fatalf("register %q: %v", take.register, err)
	}
	src, err := audio.NewFileSource(bugariPath(take.file), false, false)
	if err != nil {
		t.Skipf("recording not present: %v", err)
	}
	var all []dsp.Measurement
	eng := dsp.NewEngine(dsp.EngineConfig{
		A4:         bugariA4,
		ReedCount:  len(banks),
		Octaves:    session.OctavesOf(banks),
		FineWindow: 3 * time.Second,
	}, func(m dsp.Measurement) { all = append(all, m) })
	if err := eng.Run(context.Background(), src); err != nil {
		t.Fatalf("engine: %v", err)
	}
	return dsp.Aggregate(all, bugariA4, 0.5)
}

func TestBugariRegisters(t *testing.T) {
	if testing.Short() {
		t.Skip("loads multi-megabyte recordings")
	}
	for _, take := range bugariTakes {
		t.Run(take.file, func(t *testing.T) {
			m, ok := measureBugari(t, take)
			if !ok {
				t.Fatal("no usable reading")
			}
			if got := m.Note.Name(tuning.NamingCDEFGAB); got != take.key {
				t.Errorf("key %s, the register sounds %s", got, take.key)
			}
			if len(m.Reeds) != len(take.reeds) {
				t.Fatalf("resolved %d reeds, want %d: %+v", len(m.Reeds), len(take.reeds), m.Reeds)
			}
			for i, want := range take.reeds {
				got := m.Reeds[i]
				if math.Abs(got.Freq-want.freq) > bugariFreqTol {
					t.Errorf("reed %d at %.3f Hz, want %.3f +-%.2f", i+1, got.Freq, want.freq, bugariFreqTol)
				}
				if got.Octave != want.octave {
					t.Errorf("reed %d in band %d, want %d", i+1, got.Octave, want.octave)
				}
			}
			// A rank the engine never heard does not make the lines it did hear lobes: every one of
			// these readings is one reed per line, and a sheet that treats them as merged prints
			// nothing at all for D1 and G1.
			if !m.ReedsSeparated {
				t.Errorf("the reading reads as merged lobes (%s)", take.short)
			}
			assertBugariBands(t, m, take)
			assertBugariBeats(t, m, take)
		})
	}
}

func assertBugariBands(t *testing.T, m dsp.Measurement, take bugariTake) {
	t.Helper()
	if len(m.Bands) != len(take.bands) {
		t.Fatalf("%d bands, want %d: %+v", len(m.Bands), len(take.bands), m.Bands)
	}
	for i, want := range take.bands {
		got := m.Bands[i]
		if got.Octave != want.octave || got.Ranks != want.ranks || got.Found != want.found || got.GhostOnly != want.ghostOnly {
			t.Errorf("band %d is %+v, want octave %d ranks %d found %d ghostOnly %v (%s)",
				i, got, want.octave, want.ranks, want.found, want.ghostOnly, take.short)
		}
	}
}

func assertBugariBeats(t *testing.T, m dsp.Measurement, take bugariTake) {
	t.Helper()
	if len(m.Beats) != len(take.beats) {
		t.Fatalf("%d beats, want %d: %+v", len(m.Beats), len(take.beats), m.Beats)
	}
	for i, want := range take.beats {
		got := m.Beats[i]
		if got.Pair != want.pair {
			t.Errorf("beat %d names pair %s, want %s", i, got.Pair, want.pair)
		}
		// The sign is the reading: a rank flat of the octave beats the other way from one sharp of it.
		if math.Abs(got.Hz-want.hz) > bugariBeatTol {
			t.Errorf("beat %s at %+.3f Hz, want %+.3f +-%.2f", got.Pair, got.Hz, want.hz, bugariBeatTol)
		}
	}
}

// A beat's sign is the filing instruction: a 4' flat of the octave is scraped, one sharp of it is
// weighted. The rate may be read off the spectrum or off the band's envelope depending on how long
// the window is, and the envelope only measures how fast - so the sign used to flip with the window.
// C3 4 8's 4' rank sits 1.41 Hz under the 8's second partial at every window length.
func TestBugariBeatSignDoesNotDependOnTheWindow(t *testing.T) {
	if testing.Short() {
		t.Skip("loads multi-megabyte recordings")
	}
	take := bugariTakes[3]
	if take.file != "C3 4 8.wav" {
		t.Fatalf("expected the C3 take, got %s", take.file)
	}
	banks, err := session.ParseBanks(take.register)
	if err != nil {
		t.Fatal(err)
	}
	for _, window := range []time.Duration{1600, 2000, 2500, 3000, 3500} {
		window *= time.Millisecond
		t.Run(window.String(), func(t *testing.T) {
			src, err := audio.NewFileSource(bugariPath(take.file), false, false)
			if err != nil {
				t.Skipf("recording not present: %v", err)
			}
			var all []dsp.Measurement
			eng := dsp.NewEngine(dsp.EngineConfig{
				A4: bugariA4, ReedCount: len(banks), Octaves: session.OctavesOf(banks), FineWindow: window,
			}, func(m dsp.Measurement) { all = append(all, m) })
			if err := eng.Run(context.Background(), src); err != nil {
				t.Fatalf("engine: %v", err)
			}
			m, ok := dsp.Aggregate(all, bugariA4, 0.5)
			if !ok || len(m.Beats) != 1 {
				t.Fatalf("want one beat, got %+v", m.Beats)
			}
			if got := m.Beats[0].Hz; got > -1.3 || got < -1.5 {
				t.Errorf("beat %+.3f Hz, want about -1.41: the 4' is flat of the 8's octave "+
					"whichever way the rate was measured (fromEnvelope=%v)", got, m.Beats[0].FromEnvelope)
			}
		})
	}
}

// Every take's lowest sounding rank must be named with no register declared at all - the coarse
// detector's own job. D1 is the case that used to fail: its fundamental is 24 dB down, its 9th
// partial is the loudest line, and the tuner reported A4, the 3rd partial.
func TestBugariNoteWithoutRegister(t *testing.T) {
	if testing.Short() {
		t.Skip("loads multi-megabyte recordings")
	}
	lowest := map[string]string{
		"B1 16.wav": "A#3", "C1 16.wav": "C3", "C1 8kanal.wav": "C4",
		"C3 4 8.wav": "C6", "D1   4 16.wav": "D3",
	}
	for file, want := range lowest {
		t.Run(file, func(t *testing.T) {
			src, err := audio.NewFileSource(bugariPath(file), false, false)
			if err != nil {
				t.Skipf("recording not present: %v", err)
			}
			var all []dsp.Measurement
			eng := dsp.NewEngine(dsp.EngineConfig{
				A4: bugariA4, ReedCount: 1, FineWindow: 3 * time.Second,
			}, func(m dsp.Measurement) { all = append(all, m) })
			if err := eng.Run(context.Background(), src); err != nil {
				t.Fatalf("engine: %v", err)
			}
			m, ok := dsp.Aggregate(all, bugariA4, 0.5)
			if !ok {
				t.Fatal("no usable reading")
			}
			if got := m.Note.Name(tuning.NamingCDEFGAB); got != want {
				t.Errorf("named %s at %.2f Hz, want %s: a partial won over the reed",
					got, m.Reeds[0].Freq, want)
			}
		})
	}
}
