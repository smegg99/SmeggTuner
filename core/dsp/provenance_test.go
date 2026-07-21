package dsp

import (
	"context"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/audio"
)

// A reading says where its own audio came from. The engine measures over a three-second window and
// has already eaten audio not yet played, so a reading routinely describes a moment more than a second
// from the needle; SourceAt carries that provenance rather than letting it be guessed from the playhead.
func TestAMeasurementSaysWhereItsAudioCameFrom(t *testing.T) {
	src, err := audio.NewFileSource("../../tests/fixtures/h-16.wav", false, false)
	if err != nil {
		t.Fatal(err)
	}

	seen := make([]float64, 0, 64)
	eng := NewEngine(EngineConfig{A4: 442, ReedCount: 3}, func(m Measurement) {
		seen = append(seen, m.SourceAt)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := eng.Run(ctx, src); err != nil {
		t.Fatal(err)
	}
	if len(seen) < 10 {
		t.Fatalf("only %d measurements", len(seen))
	}

	// It only ever goes forward, and walks the recording end to end.
	var last float64
	for i, at := range seen {
		if at < last {
			t.Fatalf("measurement %d came from %.2fs, after one from %.2fs: provenance cannot go backwards", i, at, last)
		}
		last = at
	}

	first := seen[0]
	if first <= 0 || first > 1 {
		t.Fatalf("the first reading came from %.2fs, want the very start of the recording", first)
	}
	if want := src.Duration().Seconds(); last < want-0.5 || last > want+0.5 {
		t.Fatalf("the last reading came from %.2fs, but the recording is %.2fs long", last, want)
	}
}

// A microphone has no timeline, so there is nothing to point at and nothing is drawn.
func TestAMicrophoneReadingPointsNowhere(t *testing.T) {
	var m Measurement
	if m.SourceAt != 0 {
		t.Fatal("the zero measurement claims to come from somewhere")
	}

	// SynthSource stands in for the mic: neither sets Block.At, and a source with no timeline must not invent one.
	src := audio.NewSynthSource(audio.SynthSpec{
		SampleRate: 48000,
		Duration:   2 * time.Second,
		Reeds:      []audio.ReedSpec{{Freq: 440, Amp: 0.3}},
	})

	var worst float64
	eng := NewEngine(EngineConfig{A4: 440}, func(m Measurement) {
		if m.SourceAt > worst {
			worst = m.SourceAt
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := eng.Run(ctx, src); err != nil {
		t.Fatal(err)
	}
	if worst != 0 {
		t.Fatalf("a source with no timeline reported a provenance of %.2fs", worst)
	}
}
