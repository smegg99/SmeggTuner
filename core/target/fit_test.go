package target

import (
	"testing"

	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/tuning"
)

// The recorded range of the synthetic instrument: three octaves, every semitone.
const (
	fitLo = tuning.Note(48)
	fitHi = tuning.Note(84)
	fitA4 = 442.0
)

// The tremolo the instrument holds: reed 1 flat ten cents, reed 2 at pitch, reed 3 rising +5..+20.
func ramp(n tuning.Note) float64 {
	return 5 + 15*float64(n-fitLo)/float64(fitHi-fitLo)
}

func reading(n tuning.Note, devs ...float64) Reading {
	r := Reading{Note: n}
	ref := n.Freq(fitA4)
	for _, d := range devs {
		r.Reeds = append(r.Reeds, dsp.ReedMeasure{
			Freq:     tuning.FreqAtCents(ref, d),
			DevCents: d,
		})
	}
	return r
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
	res, err := Fit(rs, reedCount, fitA4, FitOptions{Name: "fitted"})
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

	if res.Used != int(fitHi-fitLo)+1 || res.Merged != 0 {
		t.Fatalf("used %d, merged %d", res.Used, res.Merged)
	}
	if len(res.Outliers) != 0 {
		t.Fatalf("a clean instrument has no outliers: %+v", res.Outliers)
	}
	for n := fitLo; n <= fitHi; n++ {
		got := res.Curve.At(n)
		almost(t, got[0], -10, 1, "reed 1 sits ten cents flat")
		almost(t, got[1], 0, 1, "reed 2 is at pitch")
		almost(t, got[2], ramp(n), 1, "reed 3 rises from +5 to +20")
	}

	// Anchors, not a dense table: it must land in the same editor as a typed curve.
	if len(res.Curve.Anchors) > 8 {
		t.Fatalf("fit produced %d anchors, which is a table and not a curve",
			len(res.Curve.Anchors))
	}
	if res.Curve.ReedCount != 3 || res.Curve.Unit != UnitCents {
		t.Fatalf("curve: %d reeds in %q", res.Curve.ReedCount, res.Curve.Unit)
	}
}

// A pass recorded one reed at a time is legal; the unrecorded reeds come back flat zero.
func TestFitOneReedPerNote(t *testing.T) {
	var rs []Reading
	for n := fitLo; n <= fitHi; n++ {
		rs = append(rs, reading(n, ramp(n)))
	}
	res := mustFit(t, rs, 3)

	for n := fitLo; n <= fitHi; n++ {
		got := res.Curve.At(n)
		almost(t, got[0], ramp(n), 1, "the reed that was recorded")
		if got[1] != 0 || got[2] != 0 {
			t.Fatalf("note %d: reeds 2 and 3 were never recorded, so their goal is "+
				"zero, not %v", n, got[1:])
		}
	}
}

// A note with fewer reeds contributes to those it carries and must not pull the missing reed to zero.
func TestFitNoteWithFewerReeds(t *testing.T) {
	rs := instrument()
	for n := fitLo; n < fitLo+12; n++ {
		rs[n-fitLo] = reading(n, -10) // the bottom octave: only reed 1 sounded
	}
	res := mustFit(t, rs, 3)

	for n := fitLo; n <= fitHi; n++ {
		got := res.Curve.At(n)
		almost(t, got[0], -10, 1, "reed 1 was recorded throughout")
		almost(t, got[2], ramp(n), 1, "reed 3 keeps its own trend where it was silent")
	}
}

// An un-separated take's peaks are lobes of one merged pair; fitting them would collapse the tremolo.
func TestFitDropsMergedTakes(t *testing.T) {
	rs := instrument()
	for n := tuning.Note(60); n <= 71; n++ {
		r := reading(n, 0, 0.4, 0.8) // lobe positions, nothing like the real reeds
		r.ReedsMerged = true
		rs[n-fitLo] = r
	}
	res := mustFit(t, rs, 3)

	if res.Merged != 12 || res.Used != int(fitHi-fitLo)+1-12 {
		t.Fatalf("used %d, merged %d", res.Used, res.Merged)
	}
	for n := fitLo; n <= fitHi; n++ {
		got := res.Curve.At(n)
		almost(t, got[0], -10, 1, "reed 1 across the merged octave")
		almost(t, got[2], ramp(n), 2, "reed 3 spans the merged octave instead of collapsing")
	}
	if len(res.Outliers) != 0 {
		t.Fatalf("a dropped take is not an outlier, it is a take: %+v", res.Outliers)
	}
}

// Every take merged: no reed pitch at all, so zero anchors (the legal no-goal curve).
func TestFitEveryTakeMerged(t *testing.T) {
	rs := instrument()
	for i := range rs {
		rs[i].ReedsMerged = true
	}
	res := mustFit(t, rs, 3)

	if res.Used != 0 || res.Merged != len(rs) {
		t.Fatalf("used %d, merged %d", res.Used, res.Merged)
	}
	if len(res.Curve.Anchors) != 0 {
		t.Fatalf("curve has %d anchors", len(res.Curve.Anchors))
	}
	for n := fitLo; n <= fitHi; n++ {
		for r, v := range res.Curve.At(n) {
			if v != 0 {
				t.Fatalf("note %d reed %d: %v", n, r, v)
			}
		}
	}
}
