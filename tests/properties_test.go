package tests

// Properties that compare the engine with itself on real audio: a lone rank, a 3-reed band, a 16'+8' register.

import (
	"errors"
	"math"
	"os"
	"reflect"
	"testing"

	"smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/dsp"
)

var propertyTakes = []struct {
	file string
	cfg  dsp.EngineConfig
}{
	{"A '8.wav", dsp.EngineConfig{A4: loneRankA4, ReedCount: 1}},
	{"hohner8+8+8A4.wav", dsp.EngineConfig{A4: 440, ReedCount: 3}},
	{"hohner16+8A4.wav", dsp.EngineConfig{A4: 440, ReedCount: 2, Octaves: []dsp.OctaveRequest{{Offset: -12, Reeds: 1}, {Offset: 0, Reeds: 1}}}},
}

// fresh runs a recording without the memo, for comparing one run against another.
func fresh(t *testing.T, file string, cfg dsp.EngineConfig) []dsp.Measurement {
	t.Helper()
	src, err := audio.NewFileSource("../sounds/"+file, false, false)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Skipf("optional recording not present: %v", err)
		}
		t.Fatalf("recording failed: %v", err)
	}
	return runSource(t, cfg, src).all
}

// A new engine per pass in a long-lived process must not depend on what played before.
func TestTheSameRecordingMeasuresTheSameTwice(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("loads recordings")
	}
	for _, take := range propertyTakes {
		t.Run(take.file, func(t *testing.T) {
			first, second := fresh(t, take.file, take.cfg), fresh(t, take.file, take.cfg)
			if len(first) != len(second) {
				t.Fatalf("%d events then %d", len(first), len(second))
			}
			for i := range first {
				if !reflect.DeepEqual(first[i], second[i]) {
					t.Fatalf("event %d differs:\n%+v\n%+v", i, first[i], second[i])
				}
			}
		})
	}
}

// The clock calibration is a pure scaling: any other change to a slot would invent a tuning error.
func TestTheCalibrationScalesEverySlotExactly(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("loads recordings")
	}
	const percent = 0.025
	for _, take := range propertyTakes {
		t.Run(take.file, func(t *testing.T) {
			shifted := take.cfg
			shifted.Calibration = percent
			plain, moved := fresh(t, take.file, take.cfg), fresh(t, take.file, shifted)
			if len(plain) != len(moved) {
				t.Fatalf("%d events uncalibrated, %d calibrated", len(plain), len(moved))
			}
			compared, skipped := 0, 0
			for i := range plain {
				a, b := plain[i], moved[i]
				if a.Frame != dsp.FrameMeasured || b.Frame != dsp.FrameMeasured || a.Note != b.Note {
					continue
				}
				for s := range a.Reeds {
					if !a.Reeds[s].Present || !b.Reeds[s].Present {
						skipped++
						continue
					}
					compared++
					if got := b.Reeds[s].Freq / a.Reeds[s].Freq; math.Abs(got-(1+percent/100)) > 1e-9 {
						t.Errorf("event %d slot %d: %.6f Hz became %.6f Hz, a factor of %.10f", i, s+1, a.Reeds[s].Freq, b.Reeds[s].Freq, got)
					}
				}
			}
			if compared == 0 {
				t.Fatal("no slot was measured under both settings")
			}
			t.Logf("%d slot readings scaled exactly, %d present under only one setting", compared, skipped)
		})
	}
}

// StopAfterLock files the free-running take, then holds its panels and measures nothing until the note stops.
func TestStopAfterLockHoldsARealTake(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("loads recordings")
	}
	for _, take := range propertyTakes {
		t.Run(take.file, func(t *testing.T) {
			gated := take.cfg
			gated.StopAfterLock = true
			free, held := fresh(t, take.file, take.cfg), fresh(t, take.file, gated)
			ref, at := takeAt(free), takeAt(held)
			if len(ref) == 0 || len(at) == 0 {
				t.Fatalf("free-running filed %d takes, gated %d", len(ref), len(at))
			}
			if at[0] != ref[0] || !reflect.DeepEqual(held[at[0]].Reeds, free[ref[0]].Reeds) {
				t.Fatalf("gating changed the take: event %d %+v, free-running event %d %+v", at[0], held[at[0]].Reeds, ref[0], free[ref[0]].Reeds)
			}
			heldFrames := 0
			for _, m := range held[at[0]+1:] {
				if m.Frame == dsp.FrameNoNote {
					break
				}
				if m.Frame == dsp.FrameNone {
					continue
				}
				if m.Frame != dsp.FrameHeld || !reflect.DeepEqual(m.Panels, held[at[0]].Panels) {
					t.Fatalf("after the take at %.2fs: %s at %.2fs with panels %+v", held[at[0]].SourceAt, m.Frame, m.SourceAt, m.Panels)
				}
				heldFrames++
			}
			t.Logf("took at %.2fs, then held %d analyses", held[at[0]].SourceAt, heldFrames)
		})
	}
}
