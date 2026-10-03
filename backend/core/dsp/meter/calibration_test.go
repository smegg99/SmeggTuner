package meter

import (
	"math"
	"testing"
)

func TestStoredCalibrationRoundsHalfAwayFromZero(t *testing.T) {
	for _, sign := range []float64{1, -1} {
		for _, c := range []struct{ in, want float64 }{{0.00249, 0.002}, {0.0025, 0.003}, {0.00251, 0.003}, {0.0625, 0.063}} {
			if got := StoredCalibration(sign * c.in); got != sign*c.want {
				t.Fatalf("StoredCalibration(%v) = %v, want %v", sign*c.in, got, sign*c.want)
			}
		}
	}
}

func TestStoredCalibrationSignedZero(t *testing.T) {
	if got := StoredCalibration(-0.0001); got != 0 || !math.Signbit(got) {
		t.Fatalf("tiny negative must store -0, got %v", got)
	}
	if got := StoredCalibration(math.Copysign(0, -1)); math.Signbit(got) {
		t.Fatal("negative zero takes the nonnegative branch")
	}
}

func TestStoredCalibrationClampsBeforeRounding(t *testing.T) {
	for _, c := range []struct{ in, want float64 }{{1e308, 500}, {-1e308, -500}, {500.0004, 500}, {math.Inf(1), 500}, {math.Inf(-1), -500}} {
		if got := StoredCalibration(c.in); got != c.want {
			t.Fatalf("StoredCalibration(%v) = %v, want %v", c.in, got, c.want)
		}
	}
	if got := StoredCalibration(math.NaN()); !math.IsNaN(got) {
		t.Fatalf("NaN passes the clamp and the rounding, got %v", got)
	}
}

func TestLoadCalibrationMissingDiffersFromZero(t *testing.T) {
	for _, parsed := range []float64{CalibrationMissing, -999998, math.Inf(-1), math.NaN()} {
		if p, ok := LoadCalibration(parsed); ok || CalibrationFactor(p, ok) != 0 {
			t.Fatalf("LoadCalibration(%v) must be missing with factor 0", parsed)
		}
	}
	if p, ok := LoadCalibration(0); !ok || p != 0 || CalibrationFactor(p, ok) != 1 {
		t.Fatalf("valid zero must give factor 1, got %v %v", p, ok)
	}
}

func TestLoadCalibrationClampsWithoutRequantizing(t *testing.T) {
	for _, c := range []struct{ in, want float64 }{{math.Nextafter(-999998, 0), -500}, {-999997, -500}, {501, 500}, {0.00249, 0.00249}} {
		if got, ok := LoadCalibration(c.in); !ok || got != c.want {
			t.Fatalf("LoadCalibration(%v) = %v %v, want %v", c.in, got, ok, c.want)
		}
	}
}

func TestCalibrationFactorDirectionAndScale(t *testing.T) {
	near := func(a, b float64) bool { return math.Abs(a-b) < 1e-9 }
	if !near(CalibrationFactor(0.003, true), 1.00003) || !near(CalibrationFactor(-0.003, true), 0.99997) {
		t.Fatal("positive percentage must raise the effective rate")
	}
	if got := 48000 * (CalibrationFactor(0.001, true) - 1); !near(got, 0.48) {
		t.Fatalf("one storage step at 48 kHz = %v Hz, want 0.48", got)
	}
}

func TestCalibrationFactorAllowsNonpositiveStorageExtremes(t *testing.T) {
	for _, c := range []struct{ in, want float64 }{{-500, -4}, {-100, 0}, {500, 6}} {
		if got := CalibrationFactor(c.in, true); got != c.want {
			t.Fatalf("CalibrationFactor(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}
