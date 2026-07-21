// core/audio/synth_test.go
package audio

import (
	"context"
	"math"
	"testing"
	"time"
)

func collect(t *testing.T, s Source) []float32 {
	t.Helper()
	ch, err := s.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var out []float32
	for b := range ch {
		out = append(out, b.Samples...)
	}
	return out
}

func TestSynthLengthAndRate(t *testing.T) {
	s := NewSynthSource(SynthSpec{
		Duration: 500 * time.Millisecond,
		Reeds:    []ReedSpec{{Freq: 440, Amp: 0.5}},
	})
	if s.Info().SampleRate != 48000 {
		t.Fatalf("default rate = %d", s.Info().SampleRate)
	}
	got := collect(t, s)
	want := 24000
	if len(got) < want-4096 || len(got) > want+4096 {
		t.Fatalf("samples = %d want ~%d", len(got), want)
	}
}

func TestSynthDeterministicAndRMS(t *testing.T) {
	spec := SynthSpec{
		Duration: 250 * time.Millisecond,
		Reeds:    []ReedSpec{{Freq: 440, Amp: 0.5}},
		NoiseAmp: 0.01,
		Seed:     7,
	}
	a := collect(t, NewSynthSource(spec))
	b := collect(t, NewSynthSource(spec))
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("not deterministic at %d", i)
		}
	}
	var sum float64
	for _, v := range a {
		sum += float64(v) * float64(v)
	}
	rms := math.Sqrt(sum / float64(len(a)))
	// sine RMS = amp/sqrt(2) ~= 0.3536, noise adds a hair
	if rms < 0.30 || rms > 0.40 {
		t.Fatalf("rms = %v", rms)
	}
}

func TestSynthHarmonicsAndHum(t *testing.T) {
	s := NewSynthSource(SynthSpec{
		Duration: 250 * time.Millisecond,
		Reeds:    []ReedSpec{{Freq: 220, Amp: 0.3, Harmonics: []float64{0.5, 0.25}}},
		HumFreq:  50,
		HumAmp:   0.05,
	})
	got := collect(t, s)
	if len(got) == 0 {
		t.Fatal("no samples")
	}
	var peak float64
	for _, v := range got {
		if a := math.Abs(float64(v)); a > peak {
			peak = a
		}
	}
	if peak > 1.0 {
		t.Fatalf("clipped: peak %v", peak)
	}
}
