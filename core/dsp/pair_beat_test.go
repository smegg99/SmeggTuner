package dsp

import (
	"math"
	"testing"
)

// The beat the technician is given, and how far it can be trusted. Read off the raw envelope it is out
// by more than a hertz on this sweep. Where the fit stands behind the pair, its beat is within a tenth
// of a hertz; where it does not, the beat falls back to the raw envelope and must not be made worse.
func TestBeatSurvivesBellows(t *testing.T) {
	lobe := 4.0 / 3.0
	worstRaw, worstSplit := 0.0, 0.0
	checked, splits := 0, 0

	for _, note := range sweepNotes {
		for _, beat := range sweepBeats {
			for _, ratio := range sweepRatios {
				for _, am := range sweepAM {
					for _, amHz := range sweepAMHz {
						zr := synthBellows(t, merged(note, beat, ratio, 0, am, amHz), note, 0.01, 3.2)
						peaks := FindPeaks(zr, 3, lobe)
						hz, _, ok := EnvelopeBeat(zr, minBeatHz, 25)
						if !ok || len(peaks) != 1 {
							continue
						}
						fit, fok := FitPair(zr, peaks[0].Freq, hz, lobe)
						got := beatOf(fit, hz)
						checked++
						if e := math.Abs(hz - beat); e > worstRaw {
							worstRaw = e
						}
						if pairVerdict(fit, fok) != "SPLIT" {
							// Refused: the beat is the envelope's, and must not be made worse.
							if math.Abs(got-hz) > 1e-9 {
								t.Errorf("%.0f Hz, beat %.2f, stroke %.2f at %.1f Hz: refused the "+
									"split but still took the model's beat (%.4f, envelope said %.4f)",
									note, beat, am, amHz, got, hz)
							}
							continue
						}
						splits++
						if e := math.Abs(got - beat); e > worstSplit {
							worstSplit = e
						}
						if math.Abs(got-beat) > 0.1 {
							t.Errorf("%.0f Hz, beat %.2f, ratio %.1f, stroke %.2f at %.1f Hz: split, "+
								"and reported the beat as %.4f (envelope alone said %.4f)",
								note, beat, ratio, am, amHz, got, hz)
						}
					}
				}
			}
		}
	}
	t.Logf("%d beats, %d of them split; the raw envelope was out by up to %.4f Hz, "+
		"the beat behind a split by up to %.4f Hz", checked, splits, worstRaw, worstSplit)

	// Teeth: if the envelope had been right all along there would be nothing to fix.
	if worstRaw <= 0.1 {
		t.Fatalf("the raw envelope was never out by more than %.4f Hz, so this test would "+
			"pass without the model doing anything", worstRaw)
	}
}
