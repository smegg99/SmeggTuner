package dsp

import (
	"math"
	"testing"
)

func sine(freq float64, sr, n int) []float32 {
	out := make([]float32, n)
	for i := range out {
		out[i] = float32(math.Sin(2 * math.Pi * freq * float64(i) / float64(sr)))
	}
	return out
}

func rms(x []float32, skip int) float64 {
	var sum float64
	for _, v := range x[skip:] {
		sum += float64(v) * float64(v)
	}
	return math.Sqrt(sum / float64(len(x)-skip))
}

func TestNotchKillsHumKeepsSignal(t *testing.T) {
	sr := 48000
	hum := sine(50, sr, sr)
	tone := sine(440, sr, sr)
	nHum := NewNotch(float64(sr), 50, 30)
	nHum.Process(hum)
	nTone := NewNotch(float64(sr), 50, 30)
	nTone.Process(tone)
	if r := rms(hum, sr/2); r > 0.02 {
		t.Fatalf("hum survived notch: rms %v", r)
	}
	if r := rms(tone, sr/2); r < 0.65 {
		t.Fatalf("tone damaged by notch: rms %v", r)
	}
}

func TestHighpassKillsThump(t *testing.T) {
	sr := 48000
	thump := sine(5, sr, sr)
	hp := NewHighpass(float64(sr), 15)
	hp.Process(thump)
	if r := rms(thump, sr/2); r > 0.05 {
		t.Fatalf("5 Hz survived highpass: rms %v", r)
	}
}
