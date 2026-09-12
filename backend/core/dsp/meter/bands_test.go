package meter

import (
	"math"
	"testing"
)

func bandsNear(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	if math.Abs(got-want) > tolerance {
		t.Fatalf("%s = %v, want %v", name, got, want)
	}
}

func TestNoteBandUnitGrid(t *testing.T) {
	b := NewNoteBand(90, 100, 110, 1000, 1000, 1)
	if b.ReferenceBin != 100 || b.NearestBin != 100 || b.LowerBin != 95 || b.UpperBin != 104 || b.BinCount() != 10 {
		t.Fatalf("band %+v count %d", b, b.BinCount())
	}
	if b.LockSpanUnits() != 40 {
		t.Fatalf("LockSpanUnits = %v", b.LockSpanUnits())
	}
	bandsNear(t, "LockSpanHz", b.LockSpanHz(), 2.4, 1e-12)
}

func TestNoteBandMidpointTiesRoundUp(t *testing.T) {
	if b := NewNoteBand(90, 101, 112, 1000, 1000, 1); b.LowerBin != 96 || b.UpperBin != 106 {
		t.Fatalf("band %+v", b)
	}
}

func TestNoteBandDoubleTransformHalvesHertzSpan(t *testing.T) {
	b := NewNoteBand(90, 100, 110, 2000, 1000, 1)
	if b.BinCount() != 20 {
		t.Fatalf("BinCount = %d", b.BinCount())
	}
	bandsNear(t, "LockSpanHz", b.LockSpanHz(), 1.2, 1e-12)
}

func TestNoteBandClockFactorChangesBinMapping(t *testing.T) {
	b := NewNoteBand(90, 100, 110, 1000, 1000, 1.25)
	if b.ReferenceBin != 80 || b.BinCount() != 8 {
		t.Fatalf("band %+v count %d", b, b.BinCount())
	}
	bandsNear(t, "LockSpanHz", b.LockSpanHz(), 3, 1e-12)
}

func TestNoteBandZeroRateFactorCountsAsOne(t *testing.T) {
	ratio := math.Pow(2, 1.0/12)
	if a, b := NewNoteBand(440/ratio, 440, 440*ratio, 524288, 48000, 0), NewNoteBand(440/ratio, 440, 440*ratio, 524288, 48000, 1); a != b {
		t.Fatalf("zero factor %+v, unit factor %+v", a, b)
	}
}

func TestNoteBandCoordinatesAgree(t *testing.T) {
	ratio := math.Pow(2, 1.0/12)
	for _, rate := range []int{44100, 48000} {
		for _, factor := range []float64{0.99997, 1, 1.00003} {
			for _, hz := range []float64{65.4, 130.8, 440, 1760} {
				b := NewNoteBand(hz/ratio, hz, hz*ratio, 524288, rate, factor)
				peak := b.ReferenceBin + 0.37
				measured := FrequencyHz(hz, DeviationUnits(peak, b.ReferenceBin))
				bandsNear(t, "measured", measured, peak*float64(rate)*factor/524288, 5e-11)
			}
		}
	}
}

func TestNoteBandEmptyBandIsPlainArithmetic(t *testing.T) {
	b := NewNoteBand(99.9, 100, 100.1, 1000, 1000, 1)
	if b.BinCount() != 0 || !math.IsInf(b.LockSpanUnits(), 1) {
		t.Fatalf("band %+v count %d span %v", b, b.BinCount(), b.LockSpanUnits())
	}
}
