package dsp

import (
	"math"
	"testing"
	"time"
)

// ringWith fills a fresh ring with a sum of sines at the given frequencies and amplitudes.
func ringWith(freqs, amps []float64, sr int, seconds float64) *Ring {
	n := int(float64(sr) * seconds)
	r := NewRing(n)
	buf := make([]float32, n)
	for i := 0; i < n; i++ {
		var v float64
		for j, f := range freqs {
			v += amps[j] * math.Sin(2*math.Pi*f*float64(i)/float64(sr))
		}
		buf[i] = float32(v)
	}
	r.Write(buf)
	return r
}

func TestZoomSinglePeak(t *testing.T) {
	r := ringWith([]float64{440.0}, []float64{0.5}, 48000, 4)
	z := NewZoom(48000)
	zr := z.Analyze(r, 440, 16, 3*time.Second)
	if !zr.Valid {
		t.Fatal("invalid")
	}
	peaks := FindPeaks(zr, 3, 0.4)
	if len(peaks) != 1 {
		t.Fatalf("peaks = %d, want 1: %+v", len(peaks), peaks)
	}
	if err := math.Abs(peaks[0].Freq - 440.0); err > 0.05 {
		t.Fatalf("freq = %v want 440 +-0.05 (err %v)", peaks[0].Freq, err)
	} else {
		t.Logf("single peak freq = %.6f Hz, err = %.6f Hz (tol 0.05)", peaks[0].Freq, err)
	}
}

func TestZoomTwoClosePeaks(t *testing.T) {
	r := ringWith([]float64{440.0, 442.6}, []float64{0.5, 0.4}, 48000, 4)
	z := NewZoom(48000)
	zr := z.Analyze(r, 441, 16, 3*time.Second)
	peaks := FindPeaks(zr, 3, 0.4)
	if len(peaks) != 2 {
		t.Fatalf("peaks = %d want 2: %+v", len(peaks), peaks)
	}
	e0 := math.Abs(peaks[0].Freq - 440.0)
	e1 := math.Abs(peaks[1].Freq - 442.6)
	if e0 > 0.05 || e1 > 0.05 {
		t.Fatalf("freqs = %v, %v (errs %v, %v)", peaks[0].Freq, peaks[1].Freq, e0, e1)
	}
	t.Logf("two peaks = %.6f, %.6f Hz, errs = %.6f, %.6f Hz (tol 0.05)",
		peaks[0].Freq, peaks[1].Freq, e0, e1)
}

func TestZoomHighNote(t *testing.T) {
	r := ringWith([]float64{3520.0}, []float64{0.3}, 48000, 4)
	z := NewZoom(48000)
	zr := z.Analyze(r, 3520, 130, 2*time.Second)
	peaks := FindPeaks(zr, 3, 0.4)
	if len(peaks) != 1 || math.Abs(peaks[0].Freq-3520.0) > 0.05 {
		t.Fatalf("peaks = %+v", peaks)
	}
	t.Logf("high note freq = %.6f Hz, err = %.6f Hz (tol 0.05)",
		peaks[0].Freq, math.Abs(peaks[0].Freq-3520.0))
}

func TestZoomInsufficientData(t *testing.T) {
	r := NewRing(48000 * 5)
	r.Write(make([]float32, 4800)) // only 0.1 s
	z := NewZoom(48000)
	if zr := z.Analyze(r, 440, 16, 3*time.Second); zr.Valid {
		t.Fatal("should be invalid with 0.1s of data")
	}
}
