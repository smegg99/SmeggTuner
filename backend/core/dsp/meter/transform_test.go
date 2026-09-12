package meter

import (
	"math"
	"math/rand/v2"
	"slices"
	"testing"
)

func transformRun(src []float32, size int) []float32 {
	dst := make([]float32, size)
	NewTransform(size).Packed(dst, src)
	return dst
}

func TestTransformFourSampleAlgebraAndPositiveImaginarySign(t *testing.T) {
	if got := transformRun([]float32{1, 2, 3, 4}, 4); !slices.Equal(got, []float32{10, -2, -2, -2}) {
		t.Fatalf("got %v", got)
	}
}

func TestTransformConstantKeepsDCAndAlternatingKeepsNyquist(t *testing.T) {
	for index, source := range [][]float32{{3, 3, 3, 3, 3, 3, 3, 3}, {3, -3, 3, -3, 3, -3, 3, -3}} {
		want := make([]float32, 8)
		want[index] = 24
		if got := transformRun(source, 8); !slices.Equal(got, want) {
			t.Fatalf("source %v: got %v", source, got)
		}
	}
}

func TestTransformShiftedImpulseMatchesDirectSum(t *testing.T) {
	src := make([]float32, 32)
	src[7] = 100
	got := transformRun(src, 32)
	for k := 1; k < 16; k++ {
		angle := 2 * math.Pi * 7 * float64(k) / 32
		if math.Abs(float64(got[2*k])-100*math.Cos(angle)) > 1e-4 || math.Abs(float64(got[2*k+1])-100*math.Sin(angle)) > 1e-4 {
			t.Fatalf("bin %d: got %v %v", k, got[2*k], got[2*k+1])
		}
	}
}

func TestTransformMatchesDirectDFT(t *testing.T) {
	r := rand.New(rand.NewPCG(3, 4))
	for _, size := range []int{8, 16, 32, 128, 512, 2048} {
		src := make([]float32, size-size/3)
		for i := range src {
			src[i] = float32(r.IntN(65536) - 32768)
		}
		got := transformRun(src, size)
		tolerance := 1e-6 * 32768 * float64(len(src))
		for k := 0; k <= size/2; k++ {
			var re, im float64
			for j, v := range src {
				angle := 2 * math.Pi * float64(j*k%size) / float64(size)
				re += float64(v) * math.Cos(angle)
				im += float64(v) * math.Sin(angle)
			}
			gotRe, gotIm := float64(got[0]), 0.0
			switch {
			case k == size/2:
				gotRe = float64(got[1])
			case k > 0:
				gotRe, gotIm = float64(got[2*k]), float64(got[2*k+1])
			}
			if math.Abs(gotRe-re) > tolerance || math.Abs(gotIm-im) > tolerance {
				t.Fatalf("size %d bin %d: got (%v, %v), want (%v, %v)", size, k, gotRe, gotIm, re, im)
			}
		}
	}
}

func TestTransformZeroPadsShortInputAndReusesThePlan(t *testing.T) {
	plan := NewTransform(64)
	short, full := []float32{5, -1, 7, 2, 9}, make([]float32, 64)
	copy(full, short)
	noise := make([]float32, 64)
	for i := range noise {
		noise[i] = float32(i * i)
	}
	a, b := make([]float32, 64), make([]float32, 64)
	plan.Packed(a, full)
	plan.Packed(b, noise)
	plan.Packed(b, short)
	if !slices.Equal(a, b) {
		t.Fatalf("padded %v, short %v", a, b)
	}
}

func TestTransformSinCosStaysWithinOneULP(t *testing.T) {
	for i := range 4096 {
		a := float64(i) * (math.Pi / 4) / 4096
		s, c := transformSinCos(a)
		if math.Abs(s-math.Sin(a)) > math.Abs(math.Nextafter(s, 2)-s) || math.Abs(c-math.Cos(a)) > math.Abs(math.Nextafter(c, 2)-c) {
			t.Fatalf("sincos(%v) = %v, %v", a, s, c)
		}
	}
}

func TestTransformRejectsUnsupportedSizes(t *testing.T) {
	for _, size := range []int{0, 2, 6, 96, 1 << 22} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatalf("size %d accepted", size)
				}
			}()
			NewTransform(size)
		}()
	}
}

func BenchmarkTransform524288(b *testing.B) {
	plan := NewTransform(1 << 19)
	r := rand.New(rand.NewPCG(1, 2))
	src := make([]float32, 1<<19)
	for i := range src {
		src[i] = float32(r.IntN(65536) - 32768)
	}
	dst := make([]float32, 1<<19)
	b.ReportAllocs()
	for b.Loop() {
		plan.Packed(dst, src)
	}
}
