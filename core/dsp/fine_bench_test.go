package dsp

import (
	"math"
	"testing"
	"time"
)

// What one fine hop costs, which decides how often the screen can be told anything.
func BenchmarkFineHop(b *testing.B) {
	const (
		sr = 48000
		fc = 440.0
	)

	ring := NewRing(sr * 4)
	buf := make([]float32, sr*4)
	for i := range buf {
		t := float64(i) / sr
		buf[i] = float32(0.4*math.Sin(2*math.Pi*441.3*t) + 0.35*math.Sin(2*math.Pi*443.1*t))
	}
	ring.Write(buf)

	zoom := NewZoom(sr)
	e := &Engine{cfg: EngineConfig{A4: 440, ReedCount: 2, FineWindow: 3 * time.Second}}
	e.cfg.fill()
	span := math.Max(16, fc*0.035)

	b.ReportAllocs()
	for b.Loop() {
		zr := zoom.Analyze(ring, fc, span, e.cfg.FineWindow)
		if !zr.Valid {
			b.Fatal("invalid zoom")
		}
		peaks := FindPeaks(zr, e.cfg.ReedCount, e.lobeWidth())
		freqs := make([]float64, len(peaks))
		for i, p := range peaks {
			freqs[i] = p.Freq
		}
		e.buildMeasurement(69, fc, freqs, peaks, zr, CoarseResult{}, [105]float64{}, 0.3, false, 0)
	}
}

// And the spectrum alone: the part moved here from the frontend.
func BenchmarkSpectrumFor(b *testing.B) {
	const (
		sr = 48000
		fc = 440.0
	)

	ring := NewRing(sr * 4)
	buf := make([]float32, sr*4)
	for i := range buf {
		buf[i] = float32(0.4 * math.Sin(2*math.Pi*441.3*float64(i)/sr))
	}
	ring.Write(buf)

	zoom := NewZoom(sr)
	e := &Engine{cfg: EngineConfig{A4: 440, ReedCount: 1, FineWindow: 3 * time.Second}}
	e.cfg.fill()
	zr := zoom.Analyze(ring, fc, math.Max(16, fc*0.035), e.cfg.FineWindow)

	b.ReportAllocs()
	for b.Loop() {
		e.specFc = 0 // the cold path: a new note, so the envelope is rebuilt
		e.spectrumFor(zr, fc)
	}
}
