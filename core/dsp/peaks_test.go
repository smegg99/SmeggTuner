package dsp

import (
	"math"
	"testing"
	"time"
)

func TestFindPeaksInvalidResult(t *testing.T) {
	if got := FindPeaks(ZoomResult{}, 3, 0.4); got != nil {
		t.Fatalf("peaks on invalid result = %+v, want nil", got)
	}
}

func TestFindPeaksMinSeparation(t *testing.T) {
	// Two candidate maxima 1 bin apart; minSepHz wider than that keeps only the stronger.
	zr := ZoomResult{
		Valid: true,
		Spec:  []float64{0, 1, 0.5, 0.9, 0, 0},
		BinHz: 0.1,
		MinHz: 439.8,
	}
	peaks := FindPeaks(zr, 3, 0.5)
	if len(peaks) != 1 {
		t.Fatalf("peaks = %+v, want 1", peaks)
	}
	if peaks[0].Amp != 1 {
		t.Fatalf("kept peak amp = %v, want the stronger (1)", peaks[0].Amp)
	}
}

func TestRefinePhaseImproves(t *testing.T) {
	// 440.013 Hz: refinement must land within 0.01 Hz
	sr := 48000
	r := NewRing(sr * 6)
	buf := make([]float32, sr*6)
	for i := range buf {
		buf[i] = float32(0.5 * math.Sin(2*math.Pi*440.013*float64(i)/float64(sr)))
	}
	// feed in two steps so we can analyze at two points in time
	r.Write(buf[:sr*5])
	z := NewZoom(sr)
	prev := z.Analyze(r, 440, 16, 3*time.Second)
	r.Write(buf[sr*5:])
	cur := z.Analyze(r, 440, 16, 3*time.Second)
	peaks := FindPeaks(cur, 1, 0.4)
	if len(peaks) != 1 {
		t.Fatalf("peaks = %+v", peaks)
	}
	refined := RefinePhase(prev, cur, peaks[0].Freq, 1.0)
	if err := math.Abs(refined - 440.013); err > 0.01 {
		t.Fatalf("refined = %v want 440.013 +-0.01 (err %v)", refined, err)
	} else {
		t.Logf("coarse = %.6f Hz (err %.6f), refined = %.6f Hz, err = %.6f Hz (tol 0.01)",
			peaks[0].Freq, math.Abs(peaks[0].Freq-440.013), refined, err)
	}
}

func TestRefinePhaseFallsBack(t *testing.T) {
	valid := ZoomResult{Valid: true, BinHz: 0.05, MinHz: 424, Center: 440,
		Phases: make([]float64, 641)}
	if got := RefinePhase(ZoomResult{}, valid, 440.0, 1.0); got != 440.0 {
		t.Fatalf("invalid prev: got %v, want passthrough 440", got)
	}
	if got := RefinePhase(valid, valid, 440.0, 0); got != 440.0 {
		t.Fatalf("zero hop: got %v, want passthrough 440", got)
	}
}

// Regression: the original expected-phase formula used (f - Center)*hop, correct only when Center*hop
// is an integer because the heterodyne LO restarts at phase zero each Analyze. The fix tracks absolute f.
func TestRefinePhaseNonIntegerCenterHop(t *testing.T) {
	sr := 48000
	r := NewRing(sr * 6)
	buf := make([]float32, sr*6)
	for i := range buf {
		buf[i] = float32(0.5 * math.Sin(2*math.Pi*440.013*float64(i)/float64(sr)))
	}
	r.Write(buf[:sr*5])
	z := NewZoom(sr)
	prev := z.Analyze(r, 440.3, 16, 3*time.Second)
	r.Write(buf[sr*5:])
	cur := z.Analyze(r, 440.3, 16, 3*time.Second)
	peaks := FindPeaks(cur, 1, 0.4)
	if len(peaks) != 1 {
		t.Fatalf("peaks = %+v", peaks)
	}
	refined := RefinePhase(prev, cur, peaks[0].Freq, 1.0)
	if math.Abs(refined-440.013) > 0.01 {
		t.Fatalf("refined = %v want 440.013 +-0.01", refined)
	}
}
