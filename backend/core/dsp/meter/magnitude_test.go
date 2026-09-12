package meter

import (
	"math"
	"slices"
	"testing"
)

func magnitudeTable(low, high float64) *NoteTable {
	var table NoteTable
	table[3], table[109] = low, high
	return &table
}

func TestMagnitudeThreeFourPairAndExclusiveMask(t *testing.T) {
	packed := make([]float32, 128)
	for k := 1; k <= 4; k++ {
		packed[2*k], packed[2*k+1] = 3, 4
	}
	got := make([]float32, 64)
	Magnitudes(got, packed, MagnitudeBounds{2, 4})
	nonzero := 0
	for _, v := range got {
		if v != 0 {
			nonzero++
		}
	}
	if got[2] != 5.0/64 || got[3] != 5.0/64 || got[4] != 0 || nonzero != 2 {
		t.Fatalf("got %v", got[:6])
	}
}

func TestMagnitudeRoundsEveryBinary32Step(t *testing.T) {
	r, s := float32(12345.678), float32(9876.543)
	packed := make([]float32, 128)
	packed[4], packed[5] = r, s
	got := make([]float32, 64)
	Magnitudes(got, packed, MagnitudeBounds{2, 3})
	want := float32(math.Sqrt(float64(float32(r*r)+float32(s*s)))) / 64
	if got[2] != want {
		t.Fatalf("got %v, want %v", got[2], want)
	}
	if want == float32(math.Hypot(float64(r), float64(s))/64) {
		t.Fatal("fixture no longer separates binary32 rounding from hypot")
	}
}

func TestMagnitudeEmptyIntervalAndAliasing(t *testing.T) {
	packed := []float32{1, 2, 3, 4, 5, 6, 7, 8}
	empty := make([]float32, 4)
	Magnitudes(empty, packed, MagnitudeBounds{2, 2})
	if !slices.Equal(empty, make([]float32, 4)) {
		t.Fatalf("empty interval wrote %v", empty)
	}
	separate := make([]float32, 4)
	Magnitudes(separate, packed, MagnitudeBounds{1, 4})
	Magnitudes(packed, packed, MagnitudeBounds{1, 4})
	if !slices.Equal(packed[:4], separate) {
		t.Fatalf("in place %v, separate %v", packed[:4], separate)
	}
}

func TestMagnitudeExceptionalValuesKeepIEEEResults(t *testing.T) {
	cases := []struct {
		value float32
		check func(float32) bool
	}{
		{float32(math.NaN()), func(v float32) bool { return v != v }},
		{float32(math.Inf(-1)), func(v float32) bool { return math.IsInf(float64(v), 1) }},
		{1e30, func(v float32) bool { return math.IsInf(float64(v), 1) }},
		{0x1p-70, func(v float32) bool { return v == 0x1p-72 }},
	}
	for _, c := range cases {
		packed := make([]float32, 8)
		packed[2] = c.value
		got := make([]float32, 4)
		Magnitudes(got, packed, MagnitudeBounds{1, 4})
		if !c.check(got[1]) {
			t.Fatalf("value %v gave %v", c.value, got[1])
		}
	}
}

func TestMagnitudePaddingScalesCoherentSinusoid(t *testing.T) {
	src := make([]float32, 32)
	for j := range src {
		src[j] = float32(400 * math.Cos(2*math.Pi*16*float64(j)/128))
	}
	packed := transformRun(src, 128)
	got := make([]float32, 64)
	Magnitudes(got, packed, MagnitudeBounds{1, 64})
	if got[16] != 100 {
		t.Fatalf("A*n/N: got %v, want 100", got[16])
	}
}

func TestMagnitudeImpulseIsFlatAndPhaseFree(t *testing.T) {
	for _, position := range []int{0, 3} {
		src := make([]float32, 32)
		src[position] = 160
		negated := make([]float32, 32)
		negated[position] = -160
		got, flipped := make([]float32, 16), make([]float32, 16)
		Magnitudes(got, transformRun(src, 32), MagnitudeBounds{1, 16})
		Magnitudes(flipped, transformRun(negated, 32), MagnitudeBounds{1, 16})
		for k := 1; k < 16; k++ {
			if math.Abs(float64(got[k])-10) > 2e-6 || got[k] != flipped[k] {
				t.Fatalf("position %d bin %d: %v, negated %v", position, k, got[k], flipped[k])
			}
		}
	}
}

func TestMagnitudeBoundsUseOuterTableEntries(t *testing.T) {
	cases := []struct {
		factor, low, high float64
		want              MagnitudeBounds
		ok                bool
	}{
		{1, 10, 50, MagnitudeBounds{10, 52}, true},
		{2, 10, 50, MagnitudeBounds{5, 26}, true},
		{1, 10, 90, MagnitudeBounds{10, 64}, true},
		{1, 5, 5, MagnitudeBounds{5, 5}, true},
		{0, 10, 50, MagnitudeBounds{}, false},
		{math.Inf(1), 10, 50, MagnitudeBounds{}, false},
		{1, 0, 50, MagnitudeBounds{}, false},
		{1, 10, 5, MagnitudeBounds{}, false},
		{1, 10, math.NaN(), MagnitudeBounds{}, false},
	}
	for _, c := range cases {
		got, ok := NewMagnitudeBounds(128, 128, c.factor, magnitudeTable(c.low, c.high))
		if ok != c.ok || (ok && got != c.want) {
			t.Fatalf("%+v: got %v %v", c, got, ok)
		}
	}
	defer func() {
		if recover() == nil {
			t.Fatal("a 127-point transform was accepted")
		}
	}()
	NewMagnitudeBounds(127, 128, 1, magnitudeTable(10, 50))
}
