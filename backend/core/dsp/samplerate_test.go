package dsp

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"smegg.me/smeggtuner/backend/core/audio"
)

func synthAt(rate int, seconds float64) *audio.SynthSource {
	return audio.NewSynthSource(audio.SynthSpec{SampleRate: rate, Duration: time.Duration(seconds * float64(time.Second)),
		Reeds: []audio.ReedSpec{{Freq: 440, Amp: 0.2}}})
}

func TestRunRefusesARateBelowTheTable(t *testing.T) {
	cases := []struct {
		rate      int
		a4        float64
		transpose int
	}{{16000, 450, 0}, {22050, 440, 11}, {32000, 450, 11}}
	for _, c := range cases {
		e := NewEngine(EngineConfig{A4: c.a4, Transpose: c.transpose}, func(Measurement) {})
		if err := e.Run(context.Background(), synthAt(c.rate, 1)); !errors.Is(err, ErrSampleRate) {
			t.Fatalf("%d Hz, A4 %v, T %d: err %v", c.rate, c.a4, c.transpose, err)
		}
	}
}

func TestRunAtTheLowestRateTheTableFits(t *testing.T) {
	cfg := EngineConfig{A4: 450, Transpose: 11}
	cfg.fill()
	tab := cfg.table()
	rate := int(math.Ceil(tab[108]+tab[109])) + 1
	e := NewEngine(cfg, func(Measurement) {})
	if err := e.Run(context.Background(), synthAt(rate, 3)); err != nil {
		t.Fatalf("%d Hz: %v", rate, err)
	}
}

func TestUpdateBeyondTheRateEndsTheRun(t *testing.T) {
	e := NewEngine(EngineConfig{}, func(Measurement) {})
	e.Update(func(c *EngineConfig) { c.Transpose = 11 })
	if err := e.Run(context.Background(), synthAt(22050, 2)); !errors.Is(err, ErrSampleRate) {
		t.Fatalf("err %v", err)
	}
}
