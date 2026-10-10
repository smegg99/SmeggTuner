// backend/core/dsp/engine_bench_test.go
package dsp

import (
	"context"
	"testing"
	"time"

	"smegg.me/smeggtuner/backend/core/audio"
)

// The whole engine over ten seconds of a three-reed note with room noise, per detection preset: what
// a second of microphone audio costs the analysis thread.
func BenchmarkEngine(b *testing.B) {
	const seconds = 10
	for _, d := range []Detection{DetectionSteady, DetectionBalanced, DetectionFast} {
		b.Run(string(d), func(b *testing.B) {
			spec := audio.SynthSpec{SampleRate: 48000, Duration: seconds * time.Second, NoiseAmp: 0.002,
				Reeds: []audio.ReedSpec{{Freq: 439.2, Amp: 0.15}, {Freq: 440, Amp: 0.15}, {Freq: 441.1, Amp: 0.15}}}
			b.ReportAllocs()
			for b.Loop() {
				e := NewEngine(EngineConfig{LearnOnStart: true, Detection: d, ReedCount: 3}, func(Measurement) {})
				if err := e.Run(context.Background(), audio.NewSynthSource(spec)); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/seconds/1e6, "ms/audio-s")
		})
	}
}

// One settings change on a running engine (a reed box, a note step, A4): the analyzer is rebuilt.
func BenchmarkReconfigure(b *testing.B) {
	cfg := EngineConfig{LearnOnStart: true, ReedCount: 3}
	cfg.fill()
	r := newRunner(cfg, 48000, false, nil)
	b.ReportAllocs()
	for b.Loop() {
		r.configure(cfg)
	}
}
