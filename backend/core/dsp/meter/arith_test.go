package meter

import (
	"math"
	"testing"
)

func TestFrequencyKeepsNativeOrder(t *testing.T) {
	if got := FrequencyHz(440, 10); got != 442.64 {
		t.Fatalf("FrequencyHz(440, 10) = %v", got)
	}
	// Sentinel-valued deviations are ordinary arithmetic; callers own missing slots.
	if got := FrequencyHz(440, -100); got != 413.6 {
		t.Fatalf("FrequencyHz(440, -100) = %v", got)
	}
}

func TestDeviationBoundariesAreExact(t *testing.T) {
	for _, c := range []struct{ bin, want float64 }{{206, 50}, {194, -50}, {188, -100}} {
		if got := DeviationUnits(c.bin, 200); got != c.want {
			t.Fatalf("DeviationUnits(%v, 200) = %v, want %v", c.bin, got, c.want)
		}
	}
}

func TestInterpolateUsesBinary32Differences(t *testing.T) {
	left, center, right := float32(1.1), float32(4.3), float32(2.7)
	rise, fall := float64(center-left), float64(center-right)
	if got := Interpolate(left, center, right); got != rise/(rise+fall)-0.5 {
		t.Fatalf("Interpolate = %v", got)
	}
	if Interpolate(1, 4, 1) != 0 {
		t.Fatal("symmetric neighbours must give zero offset")
	}
}

func TestHarmonicBeat(t *testing.T) {
	if a, b := HarmonicBeat(440, 442, 440, true), HarmonicBeat(442, 440, 440, true); a != b || math.Abs(a-2) > 1e-9 {
		t.Fatalf("unison beat %v %v", a, b)
	}
	if HarmonicBeat(0, 440, 440, true) != InvalidBeat || HarmonicBeat(440, 0, 440, false) != InvalidBeat {
		t.Fatal("zero frequency must be invalid")
	}
	if HarmonicBeat(220, 440.5, 220, false) != InvalidBeat {
		t.Fatal("octave pair has no cents beat")
	}
	if b := HarmonicBeat(220, 440.5, 220, true); math.Abs(b-0.5) > 1e-9 {
		t.Fatalf("octave Hz beat %v", b)
	}
	cents := HarmonicBeat(440, 442, 440, false)
	if want := (HarmonicBeat(440, 442, 440, true) * 100) / (440 * 0.06); cents != want {
		t.Fatalf("cents beat %v want %v", cents, want)
	}
}

func TestPairTargets(t *testing.T) {
	if TargetPair(10, 4, 440, true, false) != -6 {
		t.Fatal("signed relative difference")
	}
	if math.Abs(TargetPair(10, 4, 440, true, true)+1.584) > 1e-12 {
		t.Fatal("Hz target pair")
	}
	for _, c := range [][3]float64{{1, 2, 0}, {-99, 2, 1}, {1, -100, 1}} {
		if TargetPair(c[0], c[1], 440, c[2] == 1, true) != Missing {
			t.Fatalf("TargetPair(%v) must be Missing", c)
		}
	}
	if SubtractPairTarget(5, 3, false, false) != Missing || SubtractPairTarget(5, -99, false, true) != 5 || SubtractPairTarget(5, 3, false, true) != 2 {
		t.Fatal("pair target subtraction")
	}
	if Residual(Missing, 4) != Missing || Residual(10, Missing) != 10 || Residual(10, -100) != 110 {
		t.Fatal("residual sentinel handling")
	}
}

func TestQuantizeReadout(t *testing.T) {
	cases := []struct {
		value    float64
		decimals int
		want     float32
		plus     bool
	}{
		{0.15, 1, 0.2, true}, {-0.15, 1, -0.2, true}, {0.04, 1, 0, false}, {1.005, 2, 1.0, true}, {2.5, 0, 3, true},
	}
	for _, c := range cases {
		got, plus := QuantizeReadout(c.value, c.decimals, true)
		if got != c.want || plus != c.plus {
			t.Fatalf("QuantizeReadout(%v, %d) = %v %v, want %v %v", c.value, c.decimals, got, plus, c.want, c.plus)
		}
	}
}
