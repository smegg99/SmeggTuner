package dsp

import (
	"math"
	"testing"
	"time"
)

func TestBeatsFromPeaks(t *testing.T) {
	b := BeatsFromPeaks([]Peak{{Freq: 440.0}, {Freq: 442.6}, {Freq: 445.1}})
	if len(b) != 2 {
		t.Fatalf("len = %d", len(b))
	}
	if math.Abs(b[0].Hz-2.6) > 1e-9 || math.Abs(b[1].Hz-2.5) > 1e-9 {
		t.Fatalf("beats = %+v", b)
	}
	if b[0].FromEnvelope {
		t.Fatal("peak-derived beat must not be FromEnvelope")
	}
}

func TestEnvelopeBeatMergedReeds(t *testing.T) {
	// 0.4 Hz apart: below FindPeaks minSep, but the envelope clearly modulates at 0.4 Hz.
	r := ringWith([]float64{440.0, 440.4}, []float64{0.5, 0.5}, 48000, 6)
	z := NewZoom(48000)
	zr := z.Analyze(r, 440, 16, 5*time.Second)
	beat, depth, ok := EnvelopeBeat(zr, 0.2, 25)
	if !ok {
		t.Fatal("no envelope beat found")
	}
	// Two equal reeds swing the amplitude all the way, which tells them from one reed wobbling.
	if depth < 0.5 {
		t.Errorf("modulation depth %.2f: two equal reeds should swing deep", depth)
	}
	if math.Abs(beat-0.4) > 0.06 {
		t.Fatalf("beat = %v want 0.4 +-0.06", beat)
	}
}

func TestEnvelopeBeatSteadyTone(t *testing.T) {
	r := ringWith([]float64{440.0}, []float64{0.5}, 48000, 6)
	z := NewZoom(48000)
	zr := z.Analyze(r, 440, 16, 5*time.Second)
	if _, _, ok := EnvelopeBeat(zr, 0.2, 25); ok {
		t.Fatal("steady tone must not report an envelope beat")
	}
}
