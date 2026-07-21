package dsp

import (
	"math"
	"testing"
)

// The corner that cannot be recovered, and must therefore refuse. Stroke a pair at its own beat rate
// and the upper sideband of the lower reed lands on the upper reed: one line, which nothing in three
// seconds can split. What gives it away is the comb tooth a beat BELOW the near reed, which a pair
// leaves empty and the bellows' other sideband fills (see emptyTooth). The beat still goes out.
func TestBellowsOnTheBeatRefuses(t *testing.T) {
	lobe := 4.0 / 3.0
	for _, note := range []float64{65.4064, 130.8128} {
		for _, beat := range []float64{0.8, 1.0} {
			for _, ratio := range []float64{1.0, 0.5} {
				// The stroke sits ON the beat, within the resolution of it.
				zr := synthBellows(t, merged(note, beat, ratio, 0, 0.40, beat+0.05), note, 0.01, 3.2)
				peaks := FindPeaks(zr, 3, lobe)
				hz, _, ok := EnvelopeBeat(zr, minBeatHz, 25)
				if !ok || len(peaks) != 1 {
					continue
				}
				fit, fok := FitPair(zr, peaks[0].Freq, hz, lobe)
				if v := pairVerdict(fit, fok); v == "SPLIT" {
					t.Errorf("%.0f Hz, beat %.2f, ratio %.1f, stroke ON the beat: split into "+
						"%.4f / %.4f, and it cannot know that", note, beat, ratio, fit.Lo, fit.Hi)
				}
			}
		}
	}
}

// The wider corner around the one above, and the one that actually bites: a musette beating at 0.80 Hz
// under a forearm stroking twice a second, 0.30 Hz apart where the window resolves 0.35. The model
// widens the comb until the stroke sits a legal distance away and reports a spacing that is neither
// beat nor stroke, all of the error landing on the quiet reed. It cannot be fitted around because the
// objective PREFERS the lie (see amClear), so it is refused. Assertion is one-sided: inside amGap the
// reeds are not named. It has teeth: the same pair under a stroke clear of the beat still splits.
//
// One case inside the corner refuses nothing and is wrong anyway (a stroke at half of a LOW beat, e.g.
// 0.65 Hz pair under 0.35 Hz stroke coming back 0.616 where the truth is 0.650). It cannot be caught -
// a band with an invisible stroke is arithmetically a clean pair with a slightly different beat - so
// it is left in the open rather than hidden. It wants a longer window.
func TestBellowsNearTheBeatRefuses(t *testing.T) {
	lobe := 4.0 / 3.0

	// What a forearm on a tuning bellows does; each lands inside amGap of at least one sweep beat.
	strokes := []float64{0.5, 0.67, 0.9}

	// And one that stands well clear of every beat in the sweep.
	const clearHz = 1.7

	refused, cleared, clearSplit := 0, 0, 0
	for _, note := range sweepNotes {
		for _, beat := range sweepBeats {
			for _, ratio := range sweepRatios {
				for _, drift := range sweepDrifts {
					for _, am := range []float64{0.20, 0.40} {
						for _, fa := range strokes {
							if math.Abs(fa-beat) >= amGap {
								continue // this stroke is out in the open; the sweep has it
							}
							zr := synthBellows(t, merged(note, beat, ratio, drift, am, fa), note, 0.01, 3.2)
							peaks := FindPeaks(zr, 3, lobe)
							hz, _, ok := EnvelopeBeat(zr, minBeatHz, 25)
							if !ok || len(peaks) != 1 {
								continue
							}
							fit, fok := FitPair(zr, peaks[0].Freq, hz, lobe)
							if pairVerdict(fit, fok) != "SPLIT" {
								refused++
								continue
							}
							wantLo := note - beat/2 + drift/2
							wantHi := note + beat/2 + drift/2
							e := math.Max(math.Abs(1200*math.Log2(fit.Lo/wantLo)),
								math.Abs(1200*math.Log2(fit.Hi/wantHi)))
							t.Errorf("%.0f Hz, beat %.2f, ratio %.1f, drift %.1f, stroke %.2f at %.2f Hz "+
								"- %.2f Hz from the beat, inside the window's own gap of %.2f: split into "+
								"%.4f / %.4f (want %.4f / %.4f, %.2f cents out), and it cannot know that",
								note, beat, ratio, drift, am, fa, math.Abs(fa-beat), amGap,
								fit.Lo, fit.Hi, wantLo, wantHi, e)
						}

						// The same pair, the same depth, the stroke moved out into the open.
						if math.Abs(clearHz-beat) < amGap {
							continue
						}
						zr := synthBellows(t, merged(note, beat, ratio, drift, am, clearHz), note, 0.01, 3.2)
						peaks := FindPeaks(zr, 3, lobe)
						hz, _, ok := EnvelopeBeat(zr, minBeatHz, 25)
						if !ok || len(peaks) != 1 {
							continue
						}
						cleared++
						fit, fok := FitPair(zr, peaks[0].Freq, hz, lobe)
						if pairVerdict(fit, fok) != "SPLIT" {
							continue
						}
						clearSplit++
						wantLo := note - beat/2 + drift/2
						wantHi := note + beat/2 + drift/2
						e := math.Max(math.Abs(1200*math.Log2(fit.Lo/wantLo)),
							math.Abs(1200*math.Log2(fit.Hi/wantHi)))
						if e > 1.0 {
							t.Errorf("%.0f Hz, beat %.2f, ratio %.1f, drift %.1f, stroke %.2f at %.1f Hz, "+
								"well clear of the beat: split %.4f / %.4f, want %.4f / %.4f, %.2f cents out",
								note, beat, ratio, drift, am, clearHz, fit.Lo, fit.Hi, wantLo, wantHi, e)
						}
					}
				}
			}
		}
	}
	t.Logf("stroke inside the gap: %d refused; stroke clear of it: %d of %d split",
		refused, clearSplit, cleared)

	// Teeth: a fit that had stopped believing in bellows would refuse the clear case too.
	if clearSplit*2 < cleared {
		t.Errorf("only %d of %d split with the stroke well clear of the beat: the refusal above "+
			"is not aimed at the corner, it is the bellows being given up on", clearSplit, cleared)
	}
}

// One reed the bellows is working is not two reeds. The stroke swings loudness and pitch together,
// reinforcing one sideband and cancelling the other: a carrier and one companion. Same defect as
// TestEngineBellowsModulatedReedIsNotAPair, asserted at the fit across the whole range of pitch swing.
func TestOneWorkedReedIsNeverAPair(t *testing.T) {
	lobe := 4.0 / 3.0
	const bellows = 0.67 // the C '8 recording
	for _, am := range []float64{0.30, 0.50, 0.80} {
		for _, dev := range []float64{0.0, 0.1, 0.2, 0.3, 0.4, 0.6, 0.9} {
			ds := []bellowsReed{{f0: 262.9, amp: 0.4, am: am, amHz: bellows, fmDev: dev, fmHz: bellows}}
			zr := synthBellows(t, ds, 262.9, 0.01, 3.2)
			peaks := FindPeaks(zr, 3, lobe)
			hz, _, ok := EnvelopeBeat(zr, minBeatHz, 25)
			if !ok || len(peaks) != 1 {
				continue
			}
			fit, fok := FitPair(zr, peaks[0].Freq, hz, lobe)
			if v := pairVerdict(fit, fok); v == "SPLIT" {
				t.Errorf("one reed, loudness swung %.2f and pitch swung %.1f Hz on the same "+
					"%.2f Hz stroke: split into %.4f / %.4f at a ratio of %.2f",
					am, dev, bellows, fit.Lo, fit.Hi, fit.Ratio)
			}
		}
	}
}

// Three reeds merged into one blob are not a pair, and two of them are not the answer. The comb has a
// tooth for the third precisely so it can be seen and refused.
func TestThreeMergedReedsAreNeverAPair(t *testing.T) {
	lobe := 4.0 / 3.0
	c3 := 130.8128
	for _, third := range []float64{1.0, 0.6, 0.5, 0.35} {
		lo := c3 * math.Pow(2, -15.0/1200)
		hi := c3 * math.Pow(2, 15.0/1200)
		ds := []bellowsReed{
			{f0: lo, amp: 0.25, am: 0.1, amHz: 1.7},
			{f0: c3, amp: 0.25, am: 0.1, amHz: 1.7, ph: 1.1},
			{f0: hi, amp: 0.25 * third, am: 0.1, amHz: 1.7, ph: 2.2},
		}
		zr := synthBellows(t, ds, c3, 0.01, 3.2)
		peaks := FindPeaks(zr, 3, lobe)
		hz, _, ok := EnvelopeBeat(zr, minBeatHz, 25)
		if !ok || len(peaks) != 1 {
			continue
		}
		fit, fok := FitPair(zr, peaks[0].Freq, hz, lobe)
		if v := pairVerdict(fit, fok); v == "SPLIT" {
			t.Errorf("three reeds at -15, 0 and +15 cents (the top one at %.2f): "+
				"split into %.4f / %.4f", third, fit.Lo, fit.Hi)
		}
	}
}
