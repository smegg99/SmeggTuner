package meter

import (
	"slices"
	"testing"
)

func humFilled(count int, value float32) []float32 {
	values := make([]float32, count)
	for i := range values {
		values[i] = value
	}
	return values
}

func humRegion(t *testing.T, count int, factor float64, frequency float32) HumRegion {
	t.Helper()
	r, ok := NewHumRegion(count, 1000, factor, frequency)
	if !ok {
		t.Fatalf("no region for %d bins at %v Hz", count, frequency)
	}
	return r
}

func TestHumCenterPairAndIncludedLastPair(t *testing.T) {
	if r := humRegion(t, 64, 1, 50); r != (HumRegion{6, 1}) {
		t.Fatalf("region %v", r)
	}
	v := humFilled(64, 21)
	AttenuateHum(v, 1000, 1, 50)
	if v[5] != 1 || v[6] != 1 || v[4] != v[7] || !(v[4] > 1 && v[4] < 21) || v[3] != 21 || v[8] != 21 {
		t.Fatalf("values %v", v[2:10])
	}
}

func TestHumPairsAreSymmetricAboutHalfBin(t *testing.T) {
	for _, count := range []int{64, 1024, 262144} {
		r := humRegion(t, count, 1, 50)
		v := humFilled(count, 231)
		AttenuateHum(v, 1000, 1, 50)
		if v[r.Center] != 11 || v[r.Center+r.Width+1] != 231 {
			t.Fatalf("%d bins: center %v, outside %v", count, v[r.Center], v[r.Center+r.Width+1])
		}
		for d := range r.Width + 1 {
			if v[r.Center+d] != v[r.Center-d-1] {
				t.Fatalf("%d bins: distance %d is asymmetric", count, d)
			}
		}
	}
}

func TestHumEighthDistanceDividesByEleven(t *testing.T) {
	if r := humRegion(t, 1024, 1, 50); r != (HumRegion{102, 10}) {
		t.Fatalf("region %v", r)
	}
	v := humFilled(1024, 231)
	AttenuateHum(v, 1000, 1, 50)
	if v[110] != 21 || v[93] != 21 {
		t.Fatalf("distance eight: %v %v", v[110], v[93])
	}
}

func TestHumCalibrationMovesTheCoordinate(t *testing.T) {
	if r := humRegion(t, 1024, 2, 50); r != (HumRegion{51, 5}) {
		t.Fatalf("region %v", r)
	}
}

func TestHumOutsideSpectrumAndUnsafeLeftIndexAreNoOps(t *testing.T) {
	for _, frequency := range []float32{1, 8, 1000} {
		if _, ok := NewHumRegion(64, 1000, 1, frequency); ok {
			t.Fatalf("%v Hz gave a region", frequency)
		}
		v := humFilled(64, 3)
		AttenuateHum(v, 1000, 1, frequency)
		if !slices.Equal(v, humFilled(64, 3)) {
			t.Fatalf("%v Hz changed the spectrum", frequency)
		}
	}
}

func TestHumLeftIndexZeroIsAdmitted(t *testing.T) {
	if r := humRegion(t, 64, 1, 15.625); r != (HumRegion{2, 1}) {
		t.Fatalf("region %v", r)
	}
	v := humFilled(64, 21)
	AttenuateHum(v, 1000, 1, 15.625)
	if v[1] != 1 || v[2] != 1 || v[0] != v[3] || !(v[0] > 1 && v[0] < 21) || v[4] != 21 {
		t.Fatalf("values %v", v[:6])
	}
}

func TestHumZeroStaysZero(t *testing.T) {
	v := humFilled(64, 0)
	AttenuateHum(v, 1000, 1, 50)
	if !slices.Equal(v, humFilled(64, 0)) {
		t.Fatalf("values %v", v)
	}
}

func TestHumTwoFiltersAttenuateOverlapTwice(t *testing.T) {
	first := humFilled(1024, 231)
	AttenuateHum(first, 1000, 1, 50)
	both := slices.Clone(first)
	AttenuateHum(both, 1000, 1, 60)
	if !(both[110] < first[110]) || both[100] != first[100] || both[140] != 231 {
		t.Fatalf("overlap %v/%v, 50-only %v/%v, outside %v", both[110], first[110], both[100], first[100], both[140])
	}
}
