package dsp

import (
	"math"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/audio"
)

// The input strip draws a trace, not a bar, so every measurement carries one (heartbeats included).
func TestEngineWaveformOnEveryMeasurement(t *testing.T) {
	spec := audio.SynthSpec{
		Duration: 4 * time.Second,
		Reeds:    []audio.ReedSpec{{Freq: 440, Amp: 0.4}},
	}
	cap := runEngine(t, spec, defaultCfg())
	if len(cap.all) < 10 {
		t.Fatalf("%d measurements, want >= 10", len(cap.all))
	}

	var up, down bool
	for i, m := range cap.all {
		if len(m.Waveform) != WaveformPoints {
			t.Fatalf("measurement %d carries %d waveform points, want %d",
				i, len(m.Waveform), WaveformPoints)
		}
		for _, v := range m.Waveform {
			if v > 1 || v < -1 {
				t.Fatalf("measurement %d has %v outside -1..1", i, v)
			}
			if v > 0.5 {
				up = true
			}
			if v < -0.5 {
				down = true
			}
		}
	}
	// A sounding reed is a wave: the trace swings both ways and reaches full deflection.
	if !up || !down {
		t.Fatal("a 0.4 amplitude sine drew no swing either side of zero")
	}
}

// The trace is shape and InputLevel is loudness: a quiet room magnified to full deflection would say a reed is sounding.
func TestWaveformNormalisesButDoesNotMagnifyNoise(t *testing.T) {
	e := NewEngine(defaultCfg(), func(Measurement) {})

	// A block whose loudest sample is a fifth of full scale, the peak inside one stride so decimation keeps it.
	block := make([]float32, 4096)
	for i := range block {
		block[i] = 0.01
	}
	block[1000] = -0.2
	e.decimateWave(block)

	w := e.waveform()
	if len(w) != WaveformPoints {
		t.Fatalf("waveform = %d points, want %d", len(w), WaveformPoints)
	}
	var peak float32
	for _, v := range w {
		if a := abs32(v); a > peak {
			peak = a
		}
	}
	if math.Abs(float64(peak)-1) > 1e-6 {
		t.Fatalf("loudest point = %v, want full deflection", peak)
	}
	// Kept with its sign: a trace of magnitudes is not a waveform.
	if w[1000*WaveformPoints/len(block)] != -1 {
		t.Fatalf("the peak lost its sign: %v", w[1000*WaveformPoints/len(block)])
	}

	// Below the floor the scale is held, so noise stays small.
	quiet := make([]float32, 4096)
	for i := range quiet {
		quiet[i] = 0.005 * float32(math.Sin(float64(i)))
	}
	e.decimateWave(quiet)
	for _, v := range e.waveform() {
		if abs32(v) > 0.005/waveFloor+1e-6 {
			t.Fatalf("noise at 0.005 drew %v, want it left small", v)
		}
	}

	// Digital silence draws a flat line.
	e.decimateWave(make([]float32, 4096))
	for _, v := range e.waveform() {
		if v != 0 {
			t.Fatalf("silence drew %v", v)
		}
	}
}
