package meter

import (
	"math"
	"slices"
	"testing"
)

func streamNewest(r *SampleRing, n, capacity int) []float32 {
	dst := make([]float32, capacity)
	return dst[:r.Newest(dst, n)]
}

func TestStreamResetHistoryPrecedesFirstSamples(t *testing.T) {
	r := NewSampleRing(8)
	r.Append([]int16{10, -20, 30})
	if got := streamNewest(r, 5, 5); !slices.Equal(got, []float32{0, 0, 10, -20, 30}) || r.Appended() != 3 {
		t.Fatalf("read %v appended %d", got, r.Appended())
	}
	r.Append([]int16{32767, 32767, 32767, 32767, 32767})
	r.Reset()
	r.Append([]int16{1})
	if got := streamNewest(r, 5, 5); !slices.Equal(got, []float32{0, 0, 0, 0, 1}) || r.Appended() != 1 || r.NextWrite() != 1 {
		t.Fatalf("reset must discard loud history: %v appended %d", got, r.Appended())
	}
}

func TestStreamReadsExcludeNextWriteAndWrap(t *testing.T) {
	r := NewSampleRing(5)
	r.Append([]int16{1, 2, 3, 4})
	r.Append([]int16{5, 6, 7})
	if r.NextWrite() != 2 {
		t.Fatalf("next write %d", r.NextWrite())
	}
	if got := streamNewest(r, 5, 5); !slices.Equal(got, []float32{3, 4, 5, 6, 7}) {
		t.Fatalf("wrapped read %v", got)
	}
	if got := streamNewest(r, 2, 5); !slices.Equal(got, []float32{6, 7}) {
		t.Fatalf("contiguous read %v", got)
	}
}

func TestStreamFullCapacityWriteAtNonzeroIndex(t *testing.T) {
	r := NewSampleRing(4)
	r.Append([]int16{42})
	r.Append([]int16{-32768, -1, 0, 32767})
	if got := streamNewest(r, 4, 4); !slices.Equal(got, []float32{-32768, -1, 0, 32767}) || r.NextWrite() != 1 {
		t.Fatalf("read %v next %d", got, r.NextWrite())
	}
}

func TestStreamBoundsAndZeroRequest(t *testing.T) {
	r := NewSampleRing(4)
	r.Append([]int16{1, 2, 3, 4})
	dst := []float32{9, 9, 9, 9, 9}
	if r.Newest(dst, 0) != 0 || r.Newest(dst[:2], 3) != 0 || r.Newest(dst, 5) != 0 {
		t.Fatal("out-of-bounds requests must return zero")
	}
	if !slices.Equal(dst, []float32{9, 9, 9, 9, 9}) {
		t.Fatalf("rejected reads wrote %v", dst)
	}
	if r.Newest(dst, 3) != 3 || !slices.Equal(dst, []float32{2, 3, 4, 9, 9}) {
		t.Fatalf("read must not write past n: %v", dst)
	}
}

func TestStreamOversizedAppendMatchesSequentialWrites(t *testing.T) {
	values := []int16{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}
	whole, split := NewSampleRing(4), NewSampleRing(4)
	whole.Append([]int16{-5})
	split.Append([]int16{-5})
	whole.Append(values)
	for _, v := range values {
		split.Append([]int16{v})
	}
	if !slices.Equal(whole.samples, split.samples) || whole.NextWrite() != split.NextWrite() || whole.Appended() != 12 {
		t.Fatalf("whole %v/%d split %v/%d", whole.samples, whole.NextWrite(), split.samples, split.NextWrite())
	}
	if got := streamNewest(whole, 4, 4); !slices.Equal(got, []float32{8, 9, 10, 11}) {
		t.Fatalf("newest %v", got)
	}
}

func TestStreamDocumentedRates(t *testing.T) {
	for _, c := range []struct{ rate, coarse, fine, coarseSize, fineSize int }{
		{44100, 60211, 120422, 262144, 524288},
		{48000, 65536, 131072, 262144, 524288},
		{96000, 131072, 262144, 524288, 1048576},
		{192000, 262144, 524288, 1048576, 2097152},
	} {
		for fine, w := range [2][2]int{{c.coarse, c.coarseSize}, {c.fine, c.fineSize}} {
			p := NewWindowPlan(c.rate, fine == 1)
			if p.Samples != w[0] || p.FFTSize != w[1] || p.MagnitudeCount() != w[1]/2 {
				t.Fatalf("rate %d fine %v: %+v", c.rate, fine == 1, p)
			}
		}
	}
}

func TestStreamPowerOfTwoNeighborsAndClamps(t *testing.T) {
	for _, c := range []struct {
		rate          int
		fine          bool
		samples, size int
	}{
		{1, false, 32, 128}, {47, false, 64, 256}, {48, false, 66, 512},
		{0, false, 32, 128}, {-5, true, 32, 128}, {math.MaxInt, true, 524288, 2097152}, {400000, false, 262144, 1048576},
	} {
		p := NewWindowPlan(c.rate, c.fine)
		if p.Samples != c.samples || p.FFTSize != c.size {
			t.Fatalf("rate %d: %+v", c.rate, p)
		}
	}
}

func TestStreamWindowHasNoTaperAndDistinctZeroRegions(t *testing.T) {
	r := NewSampleRing(64)
	r.Append([]int16{12, 12, 12})
	p := NewWindowPlan(1, false)
	dst := slices.Repeat([]float32{7}, p.FFTSize+1)
	endpoint, ok := p.Read(r, dst)
	want := slices.Concat(make([]float32, 29), []float32{12, 12, 12}, make([]float32, 96), []float32{7})
	if !ok || endpoint != 3 || !slices.Equal(dst, want) {
		t.Fatalf("endpoint %d ok %v data %v", endpoint, ok, dst)
	}
}

func TestStreamReadsHaveIndependentEndpoints(t *testing.T) {
	r := NewSampleRing(64)
	r.Append([]int16{1, 2, 3})
	coarse, fine := NewWindowPlan(1, false), NewWindowPlan(1, true)
	a, b := make([]float32, coarse.FFTSize), make([]float32, fine.FFTSize)
	first, _ := coarse.Read(r, a)
	r.Append([]int16{4})
	second, _ := fine.Read(r, b)
	if first != 3 || second != 4 || a[31] != 3 || b[31] != 4 {
		t.Fatalf("endpoints %d %d samples %v %v", first, second, a[31], b[31])
	}
}

func TestStreamShortRingSkipsTheTransform(t *testing.T) {
	r := NewSampleRing(31)
	r.Append([]int16{5})
	p := NewWindowPlan(1, false)
	dst := slices.Repeat([]float32{7}, p.FFTSize)
	if _, ok := p.Read(r, dst); ok || !slices.Equal(dst, make([]float32, p.FFTSize)) {
		t.Fatalf("ok %v data %v", ok, dst)
	}
}

func TestStreamKeepZeroesAllButTheNewest(t *testing.T) {
	for _, n := range []int{0, 1, 2, 3, 5, 9} {
		r := NewSampleRing(5)
		r.Append([]int16{1, 2, 3, 4})
		r.Append([]int16{5, 6, 7})
		r.Keep(n)
		want := []float32{3, 4, 5, 6, 7}
		for i := range max(0, 5-n) {
			want[i] = 0
		}
		if got := streamNewest(r, 5, 5); !slices.Equal(got, want) || r.NextWrite() != 2 {
			t.Errorf("keep %d: %v, want %v", n, got, want)
		}
	}
}
