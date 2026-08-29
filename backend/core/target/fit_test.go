package target

import (
	"testing"

	"smegg.me/smeggtuner/backend/core/tuning"
)

// The recorded range of the synthetic instrument: three octaves, every semitone.
const (
	fitLo = tuning.Note(48)
	fitHi = tuning.Note(84)
	fitA4 = 442.0
)

// The tremolo the instrument holds: slot 1 flat ten units, slot 2 at pitch, slot 3 rising +5..+20.
func ramp(n tuning.Note) float64 {
	return 5 + 15*float64(n-fitLo)/float64(fitHi-fitLo)
}

// reading is one recorded note: a present slot per deviation, gone for a slot that did not sound.
func reading(n tuning.Note, devs ...float64) Reading {
	return measure(n, fitA4, devs...)
}

func instrument() []Reading {
	var rs []Reading
	for n := fitLo; n <= fitHi; n++ {
		rs = append(rs, reading(n, -10, 0, ramp(n)))
	}
	return rs
}

func mustFit(t *testing.T, rs []Reading, reedCount int) *FitResult {
	t.Helper()
	res, err := Fit(rs, reedCount, FitOptions{Name: "fitted"})
	if err != nil {
		t.Fatalf("fit: %v", err)
	}
	if err := res.Curve.Validate(); err != nil {
		t.Fatalf("fitted curve does not validate: %v", err)
	}
	return res
}

func TestFitRecoversTheTremoloCurve(t *testing.T) {
	res := mustFit(t, instrument(), 3)

	if res.Used != int(fitHi-fitLo)+1 {
		t.Fatalf("used %d", res.Used)
	}
	if len(res.Outliers) != 0 {
		t.Fatalf("a clean instrument has no outliers: %+v", res.Outliers)
	}
	for n := fitLo; n <= fitHi; n++ {
		got := res.Curve.At(n)
		almost(t, got[0], -10, 1, "slot 1 sits ten units flat")
		almost(t, got[1], 0, 1, "slot 2 is at pitch")
		almost(t, got[2], ramp(n), 1, "slot 3 rises from +5 to +20")
	}

	// Anchors, not a dense table: it must land in the same editor as a typed curve.
	if len(res.Curve.Anchors) > 8 {
		t.Fatalf("fit produced %d anchors, which is a table and not a curve",
			len(res.Curve.Anchors))
	}
	if res.Curve.ReedCount != 3 || res.Curve.Unit != UnitReference {
		t.Fatalf("curve: %d reeds in %q", res.Curve.ReedCount, res.Curve.Unit)
	}
}

// A pass recorded one slot at a time is legal; the unrecorded slots come back flat zero.
func TestFitOneReedPerNote(t *testing.T) {
	var rs []Reading
	for n := fitLo; n <= fitHi; n++ {
		rs = append(rs, reading(n, ramp(n)))
	}
	res := mustFit(t, rs, 3)

	for n := fitLo; n <= fitHi; n++ {
		got := res.Curve.At(n)
		almost(t, got[0], ramp(n), 1, "the slot that was recorded")
		if got[1] != 0 || got[2] != 0 {
			t.Fatalf("note %d: slots 2 and 3 were never recorded, so their goal is "+
				"zero, not %v", n, got[1:])
		}
	}
}

// A note with silent slots contributes to those it carries and must not pull the missing ones to zero.
func TestFitNoteWithFewerReeds(t *testing.T) {
	rs := instrument()
	for n := fitLo; n < fitLo+12; n++ {
		rs[n-fitLo] = reading(n, -10, gone, gone) // the bottom octave: only slot 1 sounded
	}
	res := mustFit(t, rs, 3)

	for n := fitLo; n <= fitHi; n++ {
		got := res.Curve.At(n)
		almost(t, got[0], -10, 1, "slot 1 was recorded throughout")
		almost(t, got[2], ramp(n), 1, "slot 3 keeps its own trend where it was silent")
	}
}

// Slots are positions: a silent slot in the middle does not move the slots above it down.
func TestFitKeepsSlotPositions(t *testing.T) {
	var rs []Reading
	for n := fitLo; n <= fitHi; n++ {
		rs = append(rs, reading(n, -10, gone, ramp(n)))
	}
	res := mustFit(t, rs, 3)

	for n := fitLo; n <= fitHi; n++ {
		got := res.Curve.At(n)
		almost(t, got[0], -10, 1, "slot 1")
		if got[1] != 0 {
			t.Fatalf("note %d: slot 2 never sounded, so its goal is zero, not %v", n, got[1])
		}
		almost(t, got[2], ramp(n), 1, "slot 3 stays in slot 3")
	}
}

// Nothing sounded anywhere: zero anchors, the legal no-goal curve.
func TestFitNothingSounded(t *testing.T) {
	var rs []Reading
	for n := fitLo; n <= fitHi; n++ {
		rs = append(rs, reading(n, gone, gone, gone))
	}
	res := mustFit(t, rs, 3)

	if res.Used != 0 {
		t.Fatalf("used %d silent readings", res.Used)
	}
	if len(res.Curve.Anchors) != 0 {
		t.Fatalf("curve has %d anchors", len(res.Curve.Anchors))
	}
}
