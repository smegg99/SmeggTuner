// backend/core/dsp/learned_test.go
package dsp

import (
	"context"
	"testing"

	"smegg.me/smeggtuner/backend/core/tuning"
)

func TestASecondRunStartsFromTheLearnedBackground(t *testing.T) {
	first := NewEngine(EngineConfig{LearnOnStart: true}, func(Measurement) {})
	if err := first.Run(context.Background(), scriptSource{silence: 5, tone: 8}); err != nil {
		t.Fatal(err)
	}
	learned := first.Learned()
	if learned == nil {
		t.Fatal("a finished learning pass left no background")
	}

	var ms []Measurement
	second := NewEngine(EngineConfig{LearnOnStart: true}, func(m Measurement) { ms = append(ms, m) })
	second.Seed(learned)
	if err := second.Run(context.Background(), scriptSource{silence: 1, tone: 8}); err != nil {
		t.Fatal(err)
	}
	for _, m := range ms {
		if m.Learning || m.State == StateInitializing {
			t.Fatalf("seeded run learned again: %+v", m)
		}
	}
	if m := lastMeasured(t, ms); !m.Lock.Acquired || m.Note != tuning.Note(57) {
		t.Fatalf("seeded run did not lock: %+v", m.Lock)
	}
}

func TestABackgroundFromAnotherPlanIsLearnedAgain(t *testing.T) {
	first := NewEngine(EngineConfig{LearnOnStart: true}, func(Measurement) {})
	if err := first.Run(context.Background(), scriptSource{silence: 5, tone: 1}); err != nil {
		t.Fatal(err)
	}
	var ms []Measurement
	second := NewEngine(EngineConfig{LearnOnStart: true, Hum50: true}, func(m Measurement) { ms = append(ms, m) })
	second.Seed(first.Learned())
	if err := second.Run(context.Background(), scriptSource{silence: 5, tone: 1}); err != nil {
		t.Fatal(err)
	}
	if !ms[0].Learning {
		t.Fatalf("a hum-filtered run took an unfiltered background: %+v", ms[0])
	}
}
