package tests

// sounds/bugari: a double-cassotto Bugari at A442, named by Polish octave (C1 = C4) and feet; slot frequencies are measured truth.

import (
	"math"
	"path/filepath"
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/session"
)

const (
	bugariA4      = 442.0
	bugariFreqTol = 0.05
	bugariBeatTol = 0.05
)

type bugariTake struct {
	file     string
	register string    // the banks the register sounds, as the card names them
	key      string    // the played key
	lowest   string    // the lowest sounding pitch, which a one-reed tuner names
	slots    []float64 // ground truth per slot in bank order; 0 where no reading can be checked
	octaveHz float64   // the upper of two octave slots minus twice the lower; 0 when not documented
	stray    string    // why the reference scan misreads it in file playback; see strayLine
}

var bugariTakes = []bugariTake{{
	file: "B1 16.wav", register: "L", key: "A#4", lowest: "A#3", slots: []float64{234.145},
	stray: "a 120 Hz line (A#2) scores 5-6; the 16' at A#3 scores 147-154, under 30x it, its 3rd partial F5 437-515 clears it",
}, {
	file: "C1 16.wav", register: "L", key: "C4", lowest: "C3", slots: []float64{131.286},
}, {
	file: "C1 8kanal.wav", register: "M1", key: "C4", lowest: "C4", slots: []float64{262.799},
	stray: "a 120.005 Hz line (A#2) scores 5; C4 scores 31-33 and no partial clears 30x the line, so the line itself is read",
}, {
	// The 4' sounds 1.41 Hz under the 8's second partial.
	file: "C3 4 8.wav", register: "MH", key: "C6", lowest: "C6", slots: []float64{1051.356, 2101.307}, octaveHz: -1.408,
}, {
	// The 4' is a 0.17 Hz beat off the 16's double octave, under the 2.73 s window's 0.37 Hz resolution: H reads one line for both.
	file: "D1   4 16.wav", register: "LH", key: "D4", lowest: "D3", slots: []float64{147.377, 0},
	stray: "a 120 Hz line (A#2) scores 5; D3 scores 39-46, D4 47-54 and A4 137-146, all under 30x it, the 4th partial D5 265-409 clears it",
}, {
	// The two 8' ranks are entrained into one line, so one M slot sounds and the other is not checked.
	file: "G1   4 88 16.wav", register: "LMMH", key: "G4", lowest: "G3", slots: []float64{197.093, 393.399, 0, 787.972},
	stray: "a 120 Hz line (A#2) scores 5-6; the 16' at G3 scores 30-36, under 30x it, the 8' line G4 402-470 clears it",
}}

func bugariPath(file string) string { return filepath.Join("..", "sounds", "bugari", file) }

// The register as the card declares it: the key, each slot at its bank's octave, each checked slot at its truth.
func TestBugariRegisters(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("loads recordings")
	}
	for _, take := range bugariTakes {
		t.Run(take.file, func(t *testing.T) {
			r := runFile(t, bugariPath(take.file), registerConfig(t, bugariA4, take.register))
			key, _ := heard(r.all)
			ok := name(key) == take.key && reads(r.all, key)
			if !ok && take.stray != "" {
				t.Skipf("the key is %s: %s (reads %s)", take.key, take.stray, summary(r.all))
			}
			if !ok {
				t.Fatalf("reads %s, the register sounds %s", summary(r.all), take.key)
			}
			if take.stray != "" {
				t.Errorf("listed as a stray-line misreading but reads %s: drop its stray note", take.key)
			}
			banks, _ := session.ParseBanks(take.register)
			frames := sounding(r.all, key)
			for i, want := range take.slots {
				for _, m := range frames {
					if got := m.Reeds[i].Octave; m.Reeds[i].Present && got != banks[i].Octave() {
						t.Fatalf("slot %d (%s) at octave %d, the bank sounds %d", i+1, banks[i], got, banks[i].Octave())
					}
				}
				got, n := sustained(frames, i)
				if want != 0 && (n == 0 || math.Abs(got-want) > bugariFreqTol) {
					t.Errorf("slot %d (%s) at %.3f Hz over %d of %d analyses, want %.3f +-%.2f", i+1, banks[i], got, n, len(frames), want, bugariFreqTol)
				}
				t.Logf("slot %d (%s): %.3f Hz over %d of %d analyses", i+1, banks[i], got, n, len(frames))
			}
			if take.octaveHz != 0 {
				assertOctaveBeat(t, frames, take.octaveHz)
			}
			if ts := takes(r.all); len(ts) == 0 {
				t.Logf("never locked over %d analyses on %s", len(frames), take.key)
			}
		})
	}
}

// assertOctaveBeat checks an octave pair's beat: its rate from the beat readout, its sign from the slots.
func assertOctaveBeat(t *testing.T, frames []dsp.Measurement, want float64) {
	t.Helper()
	hz, n := sustainedBeat(frames, 0)
	if n == 0 || math.Abs(hz-math.Abs(want)) > bugariBeatTol {
		t.Errorf("octave beat %.3f Hz over %d analyses, want %.3f +-%.2f", hz, n, math.Abs(want), bugariBeatTol)
	}
	lo, _ := sustained(frames, 0)
	hi, _ := sustained(frames, 1)
	if sign := hi - 2*lo; sign*want <= 0 {
		t.Errorf("the upper slot sits %+.3f Hz off the lower's octave, want the sign of %+.3f", sign, want)
	}
}

// With no register declared the plain tuner names the lowest pitch that sounds.
func TestBugariNoteWithoutRegister(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("loads recordings")
	}
	for _, take := range bugariTakes {
		t.Run(take.file, func(t *testing.T) {
			all := runFile(t, bugariPath(take.file), dsp.EngineConfig{A4: bugariA4, ReedCount: 1}).all
			got, _ := heard(all)
			ok := name(got) == take.lowest && reads(all, got)
			if !ok && take.stray != "" {
				t.Skipf("the lowest pitch is %s: %s (reads %s)", take.lowest, take.stray, summary(all))
			}
			if !ok {
				t.Errorf("reads %s, the lowest pitch is %s", summary(all), take.lowest)
			}
			if take.stray != "" && ok {
				t.Errorf("listed as a stray-line misreading but reads %s: drop its stray note", take.lowest)
			}
		})
	}
}
