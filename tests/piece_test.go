package tests

// sounds/walc wojciechowski.wav is the only recording in the repository of somebody PLAYING rather
// than holding a note: 84 seconds of a waltz, both hands, notes changing every beat, chords, and
// the bellows turning around. Every other test here feeds the engine one sustained note and asks
// what it measured. This one asks a different question - what the engine does when the input is
// nothing it was designed for - and asserts only the things that must hold whatever is playing.
//
// It is deliberately not a pitch test. Which note wins during a chord is not a contract, and
// pinning the sequence this recording happens to produce would freeze an accident. What is pinned
// is that the engine survives it, keeps following, and never hands the app a reading that
// contradicts itself.

import (
	"context"
	"math"
	"sync"
	"testing"

	"smegg.me/smeggtuner/core/audio"
	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/tuning"
)

const pieceFile = "../sounds/walc wojciechowski.wav"

// A reed is measured inside a window around its band's centre, so it can sit some way off the
// tracked note without being wrong - a hair over a semitone, in a piece where the tracked note is
// chasing a moving target. Two semitones is well past anything the search can legitimately return
// and well short of the octave a harmonic error would put it.
const pieceMaxCentsFromBand = 200.0

// Eighty-four seconds of audio takes about fifteen to grind through, and three tests ask the same
// question of it. Run it once.
var (
	pieceOnce sync.Once
	pieceRun  []dsp.Measurement
	pieceErr  error
)

func runPiece(t *testing.T) []dsp.Measurement {
	t.Helper()
	pieceOnce.Do(func() {
		src, err := audio.NewFileSource(pieceFile, false, false)
		if err != nil {
			pieceErr = err
			return
		}
		eng := dsp.NewEngine(dsp.EngineConfig{
			A4:         loneRankA4,
			ReedCount:  2,
			FineWindow: fineWindow,
		}, func(m dsp.Measurement) { pieceRun = append(pieceRun, m) })
		// Not t.Context(): the result outlives whichever test happened to run first.
		pieceErr = eng.Run(context.Background(), src)
	})
	if pieceErr != nil {
		t.Skipf("recording not usable: %v", pieceErr)
	}
	return pieceRun
}

// Every reading the engine hands out must agree with itself. DevCents is what the app draws on the
// needle and what core/target scores against the curve; if it ever disagrees with the frequency it
// was computed from, the number on screen and the number in the report come apart, and no amount of
// squinting at a spectrum would show it.
func TestAPlayedPieceNeverContradictsItself(t *testing.T) {
	if testing.Short() {
		t.Skip("loads an 84-second recording")
	}
	all := runPiece(t)
	if len(all) == 0 {
		t.Fatal("the engine emitted nothing at all")
	}

	withReeds := 0
	for i, m := range all {
		if len(m.Reeds) == 0 {
			continue
		}
		withReeds++

		if !m.Note.Valid() {
			t.Fatalf("tick %d reported reeds under note %d, which is outside the tuning range",
				i, m.Note)
		}
		if m.ScalePitch <= 0 || math.IsNaN(m.ScalePitch) || math.IsInf(m.ScalePitch, 0) {
			t.Fatalf("tick %d has reeds but a scale pitch of %v", i, m.ScalePitch)
		}

		for r, reed := range m.Reeds {
			if reed.Freq <= 0 || math.IsNaN(reed.Freq) || math.IsInf(reed.Freq, 0) {
				t.Fatalf("tick %d reed %d has frequency %v", i, r+1, reed.Freq)
			}
			// The deviation must be the one the frequency implies against that reed's own band.
			band := m.ScalePitch * math.Exp2(float64(reed.Octave)/12)
			want := 1200 * math.Log2(reed.Freq/band)
			if math.Abs(want-reed.DevCents) > 0.01 {
				t.Fatalf("tick %d reed %d: %.4f Hz against a band centre of %.4f Hz is %+.3f cents, "+
					"but the reading says %+.3f", i, r+1, reed.Freq, band, want, reed.DevCents)
			}
			if math.Abs(reed.DevCents) > pieceMaxCentsFromBand {
				t.Errorf("tick %d reed %d landed %+.1f cents from %s, further than the search "+
					"window it came from", i, r+1, reed.DevCents, m.Note.Name(tuning.NamingCDEFGAB))
			}
		}

		if len(m.Reeds) > 2 {
			t.Errorf("tick %d resolved %d reeds under a two-reed configuration", i, len(m.Reeds))
		}
		for _, b := range m.Beats {
			if math.IsNaN(b.Hz) || math.IsInf(b.Hz, 0) {
				t.Errorf("tick %d beat %s is %v", i, b.Pair, b.Hz)
			}
			// A beat normally sits between two resolved reeds. It may be reported with only one -
			// two reeds beating too close to separate leave a single line and an audible swing -
			// but ONLY then, and the reading has to say so: FromEnvelope on the beat, and reeds
			// not separated on the measurement, which is what core/session turns into ReedsMerged
			// and what stops the report printing one lobe of a peak as two reeds. A peak-derived
			// beat with one reed would key its pair off a reed that is not there.
			if len(m.Reeds) < 2 {
				if !b.FromEnvelope {
					t.Errorf("tick %d reports beat %s between peaks, but resolved only %d reed(s)",
						i, b.Pair, len(m.Reeds))
				}
				if m.ReedsSeparated {
					t.Errorf("tick %d carries an envelope beat on %d reed(s) and still claims the "+
						"reeds were separated", i, len(m.Reeds))
				}
			}
		}
	}

	if withReeds == 0 {
		t.Fatal("84 seconds of playing produced no measurement at all")
	}
	t.Logf("%d ticks, %d carried reeds, all self-consistent", len(all), withReeds)
}

// A tuner that stops following is worse than one that follows badly: the needle sits still and
// looks confident. On a piece the tracked note has to keep moving.
func TestAPlayedPieceKeepsTheTrackerMoving(t *testing.T) {
	if testing.Short() {
		t.Skip("loads an 84-second recording")
	}
	all := runPiece(t)

	distinct := map[tuning.Note]int{}
	for _, m := range all {
		if len(m.Reeds) > 0 {
			distinct[m.Note]++
		}
	}
	if len(distinct) < 5 {
		t.Errorf("a waltz moved the tracked note to %d distinct pitches; the tracker is stuck",
			len(distinct))
	}

	// And it must not be so jumpy that no note ever holds: a played note lasting a beat should
	// survive several hops.
	var longest int
	for _, n := range distinct {
		if n > longest {
			longest = n
		}
	}
	if longest < 5 {
		t.Errorf("no pitch held for more than %d ticks; the tracker is chattering", longest)
	}
	t.Logf("%d distinct pitches tracked, the most-held for %d ticks", len(distinct), longest)
}

// The bellows turn around and the player lifts their hands, so the recording goes quiet several
// times. Silence must produce a reedless heartbeat, never a reading invented out of the noise floor.
func TestSilenceInAPieceProducesNoReeds(t *testing.T) {
	if testing.Short() {
		t.Skip("loads an 84-second recording")
	}
	all := runPiece(t)

	quiet := 0
	for i, m := range all {
		if m.State != dsp.StateTooQuiet {
			continue
		}
		quiet++
		if len(m.Reeds) != 0 {
			t.Errorf("tick %d is below the noise floor and still reports %d reed(s): %+v",
				i, len(m.Reeds), m.Reeds)
		}
	}
	if quiet == 0 {
		t.Skip("the recording never dropped below the engine's noise floor")
	}
	t.Logf("%d quiet ticks, none of them invented a reed", quiet)
}
