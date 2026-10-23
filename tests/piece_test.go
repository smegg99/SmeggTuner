package tests

// An 84 s played waltz: which note wins a chord is not a contract, following and self-consistency are.

import (
	"fmt"
	"math"
	"testing"

	"smegg.me/smeggtuner/backend/core/dsp"
	"smegg.me/smeggtuner/backend/core/dsp/meter"
	"smegg.me/smeggtuner/backend/core/tuning"
)

const pieceFile = "../sounds/walc wojciechowski.wav"

func runPiece(t *testing.T) []dsp.Measurement {
	t.Helper()
	return runFile(t, pieceFile, dsp.EngineConfig{A4: loneRankA4, ReedCount: 2}).all
}

func TestAPlayedPieceNeverContradictsItself(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("loads an 84-second recording")
	}
	all := runPiece(t)
	fresh := 0
	for i, m := range all {
		if len(m.Panels) != 3 || len(m.Needles) != 2 {
			t.Fatalf("event %d carries %d panels and %d needles, the screen of two slots has 3 and 2", i, len(m.Panels), len(m.Needles))
		}
		switch m.Frame {
		case dsp.FrameMeasured, dsp.FrameIneligible:
			fresh++
			if err := consistent(m); err != nil {
				t.Fatalf("event %d (%s, %.2fs): %v", i, m.Frame, m.SourceAt, err)
			}
		case dsp.FrameNoNote:
			if !m.NeedlesDimmed || len(m.Reeds) != 0 {
				t.Fatalf("event %d has no note but %d slots and needles dimmed %v", i, len(m.Reeds), m.NeedlesDimmed)
			}
			for _, p := range m.Panels {
				if !p.Dimmed {
					t.Fatalf("event %d has no note and an undimmed panel %+v", i, p)
				}
			}
		default:
			if len(m.Reeds) != 0 || len(m.Beats) != 0 {
				t.Fatalf("event %d (%q) is not an analysis and carries %d slots", i, m.Frame, len(m.Reeds))
			}
		}
	}
	if fresh == 0 {
		t.Fatal("84 seconds of playing measured nothing")
	}
	t.Logf("%d events, %d analyses with fresh slots, all self-consistent", len(all), fresh)
}

// consistent checks one analysis against itself: key, slot frequency against deviation, missing sentinel, beat presence.
func consistent(m dsp.Measurement) error {
	if !m.Note.Valid() || m.NoteName != m.Note.Name(tuning.NamingCDEFGAB) || math.Abs(m.ScalePitch/m.Note.Freq(loneRankA4)-1) > 1e-12 {
		return fmt.Errorf("note %d %q at %v Hz", m.Note, m.NoteName, m.ScalePitch)
	}
	if len(m.Reeds) != 2 || len(m.Beats) != 1 || m.Beats[0].Pair != "1-2" {
		return fmt.Errorf("%d slots and beats %+v", len(m.Reeds), m.Beats)
	}
	for s, r := range m.Reeds {
		if !r.Present {
			if r.Dev != meter.Missing || r.Freq != 0 {
				return fmt.Errorf("slot %d missing but %+v", s+1, r)
			}
			continue
		}
		nominal := m.ScalePitch * math.Exp2(float64(r.Octave)/12)
		if want := nominal * (1 + 0.0006*r.Dev); r.Freq <= 0 || math.Abs(r.Freq/want-1) > 1e-12 {
			return fmt.Errorf("slot %d at %v Hz is not %+.3f units from %.4f Hz", s+1, r.Freq, r.Dev, nominal)
		}
		if m.Frame == dsp.FrameMeasured && math.Abs(r.Dev) > 58 {
			return fmt.Errorf("slot %d at %+.1f units is beyond the display range on a measured frame", s+1, r.Dev)
		}
	}
	b := m.Beats[0]
	if b.Present != (m.Reeds[0].Present && m.Reeds[1].Present) || b.Present && !(b.Hz >= 0) {
		return fmt.Errorf("beat %+v over %+v", b, m.Reeds)
	}
	return nil
}

// A tuner that stops following is worse than one that follows badly; one that never holds a note is useless.
func TestAPlayedPieceKeepsTheTrackerMoving(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("loads an 84-second recording")
	}
	distinct := map[tuning.Note]bool{}
	longest, held, last := 0, 0, tuning.Note(0)
	for _, m := range measured(runPiece(t)) {
		distinct[m.Note] = true
		if m.Note != last {
			last, held = m.Note, 0
		}
		held++
		longest = max(longest, held)
	}
	if len(distinct) < 5 {
		t.Errorf("a waltz moved the note to %d distinct pitches: the tracker is stuck", len(distinct))
	}
	if longest < 5 {
		t.Errorf("no note held for more than %d consecutive analyses: the tracker is chattering", longest)
	}
	t.Logf("%d distinct notes, the longest held for %d consecutive analyses", len(distinct), longest)
}
