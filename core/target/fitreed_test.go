package target

import (
	"math"
	"testing"

	"smegg.me/smeggtuner/core/tuning"
)

// scatter gives the instrument the spread of a real accordion, so a fit is not
// tested only on readings that sit exactly on the curve.
func scatter(n tuning.Note, reed int) float64 {
	return 2.5 * math.Sin(float64(int(n)*7+reed*31))
}

func scattered(drift func(tuning.Note) float64) []Reading {
	var rs []Reading
	for n := fitLo; n <= fitHi; n++ {
		rs = append(rs, reading(n,
			-10+scatter(n, 0),
			0+scatter(n, 1),
			ramp(n)+scatter(n, 2)+drift(n)))
	}
	return rs
}

// A badly tuned reed must not drag the curve towards itself: bend it ten cents
// and the app would report the bad reed thirty out and its neighbours ten out.
func TestFitIsNotDraggedByOneBadReed(t *testing.T) {
	const bad = tuning.Note(60)

	rs := instrument()
	rs[bad-fitLo] = reading(bad, -10, 0, 40)
	res := mustFit(t, rs, 3)

	for n := fitLo; n <= fitHi; n++ {
		almost(t, res.Curve.At(n)[2], ramp(n), 2, "the curve still describes the instrument")
	}

	if len(res.Outliers) != 1 {
		t.Fatalf("want exactly the one bad reed reported, got %+v", res.Outliers)
	}
	o := res.Outliers[0]
	if o.Note != bad || o.Reed != 2 {
		t.Fatalf("outlier is note %d reed %d", o.Note, o.Reed)
	}
	almost(t, o.Curr, 40, 1e-9, "the outlier reports what was recorded")
	almost(t, o.Goal, ramp(bad), 2, "and what the curve asks of it")
	almost(t, o.Error, 40-ramp(bad), 2, "so the technician knows what to take off")

	// The neighbours are the real casualties of a fit that bends; they stay in tol.
	for _, n := range []tuning.Note{bad - 1, bad + 1} {
		e := Errors(measure(n, fitA4, -10, 0, ramp(n)), res.Curve, fitA4, 0)
		if !e[2].InTol {
			t.Fatalf("note %d reed 3 reads %.2f cents out because of its bad neighbour",
				n, e[2].Error)
		}
	}
}

// The case the median is for: a run of reeds in one octave all drifted the same
// way. An average moves with them and takes the rejection window along; a median
// does not move until half the octave has gone.
func TestFitSurvivesADriftedBlock(t *testing.T) {
	bad := []tuning.Note{60, 61, 62, 63, 64}

	rs := instrument()
	for _, n := range bad {
		rs[n-fitLo] = reading(n, -10, 0, ramp(n)+12)
	}
	res := mustFit(t, rs, 3)

	for n := fitLo; n <= fitHi; n++ {
		almost(t, res.Curve.At(n)[2], ramp(n), 1,
			"the drifted block must not become the instrument's tremolo")
	}
	if len(res.Outliers) != len(bad) {
		t.Fatalf("want the %d drifted reeds reported, got %+v", len(bad), res.Outliers)
	}
	for i, o := range res.Outliers {
		if o.Note != bad[i] || o.Reed != 2 {
			t.Fatalf("outlier %d is note %d reed %d", i, o.Note, o.Reed)
		}
		almost(t, o.Error, 12, 1, "and each of them is twelve cents sharp")
	}

	// The fine notes of that octave must keep reading as fine.
	for n := tuning.Note(65); n <= 71; n++ {
		e := Errors(measure(n, fitA4, -10, 0, ramp(n)), res.Curve, fitA4, 0)
		if !e[2].InTol {
			t.Fatalf("note %d reed 3 is in tune and reads %.2f cents out", n, e[2].Error)
		}
	}
}

// A third of one rank drifted sharp, spread across the keyboard but under half of
// every octave. Measured on this pass: an average bends 3.2 cents and reports 17
// bad reeds when 13 are; the median bends 1.4 and reports the 13.
func TestFitSurvivesAWidelyDriftedRank(t *testing.T) {
	drifted := func(n tuning.Note) bool { return int(n)%3 == 0 }
	rs := scattered(func(n tuning.Note) float64 {
		if drifted(n) {
			return 12
		}
		return 0
	})
	res := mustFit(t, rs, 3)

	for n := fitLo; n <= fitHi; n++ {
		almost(t, res.Curve.At(n)[2], ramp(n), 2,
			"the drifted reeds must not become the instrument's tremolo")
	}

	// Exactly the 13, not "at least": flagging an honest reed is the same lie the
	// other way round.
	var want []tuning.Note
	for n := fitLo; n <= fitHi; n++ {
		if drifted(n) {
			want = append(want, n)
		}
	}
	if len(res.Outliers) != len(want) {
		t.Fatalf("want the %d drifted reeds and nothing else, got %d",
			len(want), len(res.Outliers))
	}
	for i, o := range res.Outliers {
		if o.Note != want[i] || o.Reed != 2 {
			t.Fatalf("outlier %d is note %d reed %d, want note %d reed 3",
				i, o.Note, o.Reed+1, want[i])
		}
	}
}

// An octave holding a single reading cannot median anything: its anchor is that
// reading and the curve passes through it, so one bad reed at the top of the
// keyboard used to define the top of the curve silently.
func TestFitThinOctaveCannotDefineTheCurve(t *testing.T) {
	rs := instrument() // note 84 is alone in its octave
	rs[fitHi-fitLo] = reading(fitHi, -10, 0, ramp(fitHi)+25)
	res := mustFit(t, rs, 3)

	almost(t, res.Curve.At(fitHi)[2], ramp(fitHi), 2,
		"the top of the curve is the instrument's, not one bad reed's")
	if len(res.Outliers) != 1 {
		t.Fatalf("want the one bad reed reported, got %+v", res.Outliers)
	}
	if o := res.Outliers[0]; o.Note != fitHi || o.Reed != 2 {
		t.Fatalf("outlier is note %d reed %d", o.Note, o.Reed+1)
	}
	almost(t, res.Outliers[0].Error, 25, 2, "and it is twenty five cents sharp")
}
