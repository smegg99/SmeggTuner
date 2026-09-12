package meter

import (
	"math"
	"slices"
	"testing"
)

func TestCaptureThresholdsAreBinary32Ratios(t *testing.T) {
	if math.Float32bits(captureLeftQuiet) != 0x3b800100 || math.Float32bits(captureRightActive) != 0x3c800100 {
		t.Fatalf("thresholds %#x %#x", math.Float32bits(captureLeftQuiet), math.Float32bits(captureRightActive))
	}
}

func TestCaptureFloatClampsAndKeepsSign(t *testing.T) {
	values := []float32{-3, -1, -0.5, float32(math.Copysign(0, -1)), 0, 0.5, 1, 3}
	chunk := captureRun(t, captureFloats(values...), len(values), 1, CaptureFloat32, 0)
	captureExpect(t, chunk, 0, -32767, -32767, -16384, 0, 0, 16384, 32767, 32767)
	if !chunk.TimestampEligible || chunk.Silent {
		t.Fatalf("plain chunk flags: %+v", chunk)
	}
}

func TestCaptureFloatThresholdNeighbours(t *testing.T) {
	for _, c := range []struct {
		left, right uint32
		channel     int
		sample      int16
	}{{0x3b800100, 0x3c800100, 1, 512}, {0x3b800101, 0x3c800100, 0, 128}, {0x3b800100, 0x3c8000ff, 0, 128}} {
		for _, sign := range []float32{-1, 1} {
			data := captureFloats(sign*math.Float32frombits(c.left), sign*math.Float32frombits(c.right))
			captureExpect(t, captureRun(t, data, 1, 2, CaptureFloat32, 0), c.channel, int16(sign)*c.sample)
		}
	}
}

func TestCaptureStereoSelectsPerChunkWithoutMixing(t *testing.T) {
	captureExpect(t, captureRun(t, captureFloats(0.5, -0.5, -0.5, 0.5), 2, 2, CaptureFloat32, 0), 0, 16384, -16384)
	captureExpect(t, captureRun(t, captureFloats(0.01, 1), 1, 2, CaptureFloat32, 0), 0, 328)
	data := captureFloats(0, 0.5, 0.25, 0.75)
	captureExpect(t, captureRun(t, data, 2, 2, CaptureFloat32, 0), 0, 0, 8192)
	captureExpect(t, captureRun(t, data[:8], 1, 2, CaptureFloat32, 0), 1, 16384)
	captureExpect(t, captureRun(t, data[8:], 1, 2, CaptureFloat32, 0), 0, 8192)
}

func TestCaptureRoundsTheBinary32Product(t *testing.T) {
	// 0x37800100 * 32767 is below 0.5 exactly but rounds to 0.5 in binary32.
	at, below := math.Float32frombits(0x37800100), math.Float32frombits(0x378000ff)
	captureExpect(t, captureRun(t, captureFloats(at, -at, below, -below), 4, 1, CaptureFloat32, 0), 0, 1, -1, 0, 0)
}

func TestCaptureFloatIsMonotoneAndOdd(t *testing.T) {
	var positive, negative []float32
	for i := range 4097 {
		positive, negative = append(positive, float32(i)/4096), append(negative, -float32(i)/4096)
	}
	up := captureRun(t, captureFloats(positive...), len(positive), 1, CaptureFloat32, 0).Samples
	down := captureRun(t, captureFloats(negative...), len(negative), 1, CaptureFloat32, 0).Samples
	if !slices.IsSorted(up) || up[0] != 0 || up[len(up)-1] != 32767 {
		t.Fatalf("grid not monotone from 0 to 32767: %v ... %v", up[:3], up[len(up)-3:])
	}
	for i := range up {
		if down[i] != -up[i] {
			t.Fatalf("grid %d: %d is not -%d", i, down[i], up[i])
		}
	}
}

func TestCaptureFloatExtremesAndNaN(t *testing.T) {
	tiny, huge := math.Float32frombits(0x00800000), math.Float32frombits(0x7f7fffff)
	captureExpect(t, captureRun(t, captureFloats(tiny, -tiny, huge, -huge), 4, 1, CaptureFloat32, 0), 0, 0, 0, 32767, -32767)
	nan, inf := float32(math.NaN()), float32(math.Inf(1))
	data := captureFloats(nan, -nan, inf, -inf, math.Float32frombits(1), -math.Float32frombits(1))
	captureExpect(t, captureRun(t, data, 6, 1, CaptureFloat32, 0), 0, 0, 0, 32767, -32767, 0, 0)
}

func TestCaptureNaNNeverRaisesAPeak(t *testing.T) {
	nan := float32(math.NaN())
	captureExpect(t, captureRun(t, captureFloats(nan, 1), 1, 2, CaptureFloat32, 0), 1, 32767)
	captureExpect(t, captureRun(t, captureFloats(0, nan), 1, 2, CaptureFloat32, 0), 0, 0)
	captureExpect(t, captureRun(t, captureFloats(0.5, 1, nan, 1), 2, 2, CaptureFloat32, 0), 0, 16384, 0)
}
