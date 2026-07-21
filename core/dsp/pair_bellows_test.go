package dsp

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

// The fine stage against the condition every recording is made in: a tuning bellows pumped by hand.
// The stroke amplitude-modulates both reeds of a pair, throwing a sideband a stroke either side of
// each (13-28% of the carrier on real recordings), and it lands in the same 1.33 Hz main lobe as the
// beat, so the raw envelope blends the two - a pair beating at 1.00 Hz under a 1.7 Hz stroke read as 1.54.

// bellowsReed is one reed of the sweep: a pitch that slides, a loudness the bellows swings, and a phase.
type bellowsReed struct {
	f0, drift   float64
	fmDev, fmHz float64
	amp, am     float64
	amHz, ph    float64
}

// synthBellows renders reeds to a zoomed band the way the engine sees one.
func synthBellows(t *testing.T, ds []bellowsReed, centre, noise, secs float64) ZoomResult {
	t.Helper()
	sr := 48000
	n := int(float64(sr) * secs)
	buf := make([]float32, n)
	rng := rand.New(rand.NewSource(11))
	phases := make([]float64, len(ds))
	for i := 0; i < n; i++ {
		ts := float64(i) / float64(sr)
		frac := ts / secs
		var v float64
		for j, d := range ds {
			f := d.f0 + d.drift*frac + d.fmDev*math.Sin(2*math.Pi*d.fmHz*ts+d.ph)
			phases[j] += 2 * math.Pi * f / float64(sr)
			a := d.amp * (1 + d.am*math.Sin(2*math.Pi*d.amHz*ts+d.ph))
			v += a * math.Sin(phases[j])
		}
		if noise > 0 {
			v += noise * (2*rng.Float64() - 1)
		}
		buf[i] = float32(v)
	}
	ring := NewRing(n)
	ring.Write(buf)
	return NewZoom(sr).Analyze(ring, centre, math.Max(16, centre*0.035), 3*time.Second)
}

// pairVerdict is splitPair's three tests and nothing else, so this file and the engine cannot drift
// on what counts as a pair.
func pairVerdict(fit PairFit, ok bool) string {
	switch {
	case !ok:
		return "refuse(fit)"
	case !fit.Split:
		return "refuse(shape)"
	case fit.Explained < pairExplained:
		return "refuse(explained)"
	case fit.Depth < PairDepth(fit.Ratio)-pairDepthSlack:
		return "refuse(depth)"
	}
	return "SPLIT"
}

// merged is a pair too close for the spectrum to separate, on a working bellows, with its harmonics,
// drift and noise. This is the shape of every case below.
func merged(note, beat, ratio, drift, am, amHz float64) []bellowsReed {
	return []bellowsReed{
		{f0: note - beat/2, drift: drift, amp: 0.25, am: am, amHz: amHz},
		{f0: note + beat/2, drift: drift, amp: 0.25 * ratio, am: am, amHz: amHz, ph: 1.0},
		{f0: 2 * (note - beat/2), drift: 2 * drift, amp: 0.075, am: am, amHz: amHz},
		{f0: 2 * (note + beat/2), drift: 2 * drift, amp: 0.075 * ratio, am: am, amHz: amHz, ph: 1.0},
	}
}

// The registers a musette lives in: a constant-cents tremolo shrinks in hertz as notes fall.
var (
	sweepNotes  = []float64{65.4064, 130.8128, 261.6256} // C2, C3, C4
	sweepBeats  = []float64{0.65, 0.8, 1.0, 1.2}
	sweepRatios = []float64{1.0, 0.7, 0.5, 0.4}
	sweepDrifts = []float64{0, 0.3}
	sweepAM     = []float64{0.0, 0.20, 0.40}
	sweepAMHz   = []float64{0.5, 0.9, 1.7}
)

// Every pair this reports as split is within a cent of both its reeds, at every stroke depth. The
// assertion is one-sided and on the ANSWER, not the yield: a pair the fit will not stand behind is
// reported as a beat alone. The yield is logged (63% of the sweep at time of writing), not asserted,
// but a fit that refused everything would be useless, so a floor is checked.
func TestFitPairUnderBellows(t *testing.T) {
	lobe := 4.0 / 3.0
	type tally struct{ fitted, split int }
	byAM := map[float64]*tally{}
	for _, am := range sweepAM {
		byAM[am] = &tally{}
	}
	worst := 0.0

	for _, note := range sweepNotes {
		for _, beat := range sweepBeats {
			for _, ratio := range sweepRatios {
				for _, drift := range sweepDrifts {
					for _, am := range sweepAM {
						for _, amHz := range sweepAMHz {
							zr := synthBellows(t, merged(note, beat, ratio, drift, am, amHz), note, 0.01, 3.2)
							peaks := FindPeaks(zr, 3, lobe)
							hz, _, ok := EnvelopeBeat(zr, minBeatHz, 25)
							if !ok || len(peaks) != 1 {
								continue
							}
							fit, fok := FitPair(zr, peaks[0].Freq, hz, lobe)
							byAM[am].fitted++
							if pairVerdict(fit, fok) != "SPLIT" {
								continue
							}
							byAM[am].split++

							// A reed that slid across the window is measured at its mean.
							wantLo := note - beat/2 + drift/2
							wantHi := note + beat/2 + drift/2
							e := math.Max(math.Abs(1200*math.Log2(fit.Lo/wantLo)),
								math.Abs(1200*math.Log2(fit.Hi/wantHi)))
							if e > worst {
								worst = e
							}
							if e > 1.0 {
								t.Errorf("%.0f Hz, beat %.2f, ratio %.1f, drift %.1f, stroke %.2f at %.1f Hz: "+
									"split %.4f / %.4f, want %.4f / %.4f, %.2f cents out",
									note, beat, ratio, drift, am, amHz,
									fit.Lo, fit.Hi, wantLo, wantHi, e)
							}
						}
					}
				}
			}
		}
	}

	var fitted, split int
	for _, am := range sweepAM {
		fitted += byAM[am].fitted
		split += byAM[am].split
		t.Logf("stroke %.2f: %d of %d split (%.0f%%)", am, byAM[am].split, byAM[am].fitted,
			100*float64(byAM[am].split)/float64(byAM[am].fitted))
	}
	t.Logf("all: %d of %d split (%.0f%%); worst error of a split %.3f cents",
		split, fitted, 100*float64(split)/float64(fitted), worst)

	if split*2 < fitted {
		t.Errorf("only %d of %d split: the bellows has cost more than it should", split, fitted)
	}
}
