package tuner

import (
	"testing"

	"smegg.me/smeggtuner/core/dsp"
)

// reading is a fine result: it carries a note.
func reading(locked bool) dsp.Measurement {
	return dsp.Measurement{
		Note:       69,
		ScalePitch: 440,
		Locked:     locked,
		Reeds:      []dsp.ReedMeasure{{Freq: 440.6, DevCents: 2.4}},
		Equalizer:  []float32{1, 2, 3},
		InputLevel: 0.2,
		State:      dsp.StateRunning,
	}
}

// heartbeat is what core/dsp emits between fine results: meters, and no reading.
func heartbeat() dsp.Measurement {
	return dsp.Measurement{Equalizer: []float32{1, 2, 3}, InputLevel: 0.2, State: dsp.StateRunning}
}

func held(m dsp.Measurement) bool { return len(m.Reeds) == 0 }

// A held reading is not a dead engine: its meters keep moving.
func TestHeldReadingKeepsTheMeters(t *testing.T) {
	var h holder
	r := rules{stopAfterLock: true}

	if got := h.filter(reading(true), r); held(got) {
		t.Fatal("the first locked reading is the one he asked to look at: it must land")
	}
	got := h.filter(reading(true), r)
	if !held(got) {
		t.Fatal("the second locked reading must not repaint: that is the whole switch")
	}
	if got.InputLevel == 0 || len(got.Equalizer) == 0 || got.State != dsp.StateRunning {
		t.Fatalf("a held reading must keep its meters, got %+v", got)
	}
	if got.Note != 0 || got.ScalePitch != 0 || got.Locked {
		t.Fatalf("a held reading must carry no reading, got %+v", got)
	}
}

// The hold releases when the note does; without it the tuner would show one reading forever.
func TestStopAfterLockReleasesOnUnlock(t *testing.T) {
	var h holder
	r := rules{stopAfterLock: true}

	h.filter(reading(true), r)
	if !held(h.filter(reading(true), r)) {
		t.Fatal("still locked: the reading holds")
	}
	if held(h.filter(reading(false), r)) {
		t.Fatal("the note was released: the tuner is live again")
	}
	if held(h.filter(reading(true), r)) {
		t.Fatal("the next note locked: its reading must land")
	}
}

// A heartbeat says nothing about the lock and must never be read as an unlock.
func TestHeartbeatDoesNotReleaseTheHold(t *testing.T) {
	var h holder
	r := rules{stopAfterLock: true}

	h.filter(reading(true), r)
	if got := h.filter(heartbeat(), r); held(got) != true || got.InputLevel == 0 {
		t.Fatalf("a heartbeat travels as it is, got %+v", got)
	}
	if !held(h.filter(reading(true), r)) {
		t.Fatal("a heartbeat between two locked readings must not release the hold")
	}
}

// With switches off, every fine result repaints.
func TestSwitchesOffChangeNothing(t *testing.T) {
	var h holder
	for _, r := range []rules{{}, {manual: true, continuousManual: true}} {
		for _, locked := range []bool{true, true, false, true} {
			if held(h.filter(reading(locked), r)) {
				t.Fatalf("rules %+v held a reading it was never asked to hold", r)
			}
		}
	}
}

// The pinned-note switch only: with it off, an unlocked (still-settling) reading does not repaint.
func TestContinuousUpdateManual(t *testing.T) {
	cases := []struct {
		name       string
		r          rules
		wantUnlock bool // an unlocked reading repaints
	}{
		{"manual, not continuous", rules{manual: true}, false},
		{"manual, continuous", rules{manual: true, continuousManual: true}, true},
		{"auto, not continuous", rules{}, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var h holder
			if held(h.filter(reading(false), c.r)) == c.wantUnlock {
				t.Fatalf("unlocked reading repainted = %v, want %v", !c.wantUnlock, c.wantUnlock)
			}
			// A locked reading lands whatever this switch says.
			if held(h.filter(reading(true), c.r)) {
				t.Fatal("a locked reading must always land")
			}
		})
	}
}

// The two switches compose: in manual mode an unlocked reed does not repaint and the locked reading holds.
func TestBothSwitchesCompose(t *testing.T) {
	var h holder
	r := rules{stopAfterLock: true, continuousManual: false, manual: true}

	if !held(h.filter(reading(false), r)) {
		t.Fatal("manual and not continuous: an unlocked reading does not repaint")
	}
	if held(h.filter(reading(true), r)) {
		t.Fatal("the lock is the reading he is waiting for: it must land")
	}
	if !held(h.filter(reading(true), r)) {
		t.Fatal("and then it holds still")
	}
}
