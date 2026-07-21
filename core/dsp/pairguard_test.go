package dsp

import (
	"context"
	"math"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/audio"
	"smegg.me/smeggtuner/core/tuning"
)

// Below the envelope's floor the pair stays merged and the engine says nothing about the reeds inside
// it: a swing that slow is the bellows breathing, and nothing is invented to fill the gap.
func TestEngineBelowBeatFloorInventsNothing(t *testing.T) {
	centre := noteC3.Freq(440)
	cfg := defaultCfg()
	cfg.ReedCount = 2
	cap := runEngine(t, pairSpec(centre, 0.4, 1.0, 12), cfg)

	cap.mu.Lock()
	defer cap.mu.Unlock()
	for i, m := range cap.all {
		if m.ReedsFromBeat {
			t.Fatalf("tick %d: split a pair beating at 0.4 Hz, under the %.1f Hz floor: %+v",
				i, minBeatHz, m.Reeds)
		}
		if len(m.Reeds) > 1 {
			t.Fatalf("tick %d: %d reeds out of a pair that cannot be resolved: %+v",
				i, len(m.Reeds), m.Reeds)
		}
		if len(m.Beats) > 0 {
			t.Fatalf("tick %d: beat under the floor: %+v", i, m.Beats)
		}
	}
}

// Three reeds merged into one blob (a musette at C3) are not a pair: two would be right and the third
// simply gone, and a technician reading two reeds where three sound has been lied to.
func TestEngineThreeMergedReedsAreNotAPair(t *testing.T) {
	centre := noteC3.Freq(440)
	spec := audio.SynthSpec{
		Duration: 10 * time.Second,
		Reeds: []audio.ReedSpec{
			{Freq: tuning.FreqAtCents(centre, -15), Amp: 0.4},
			{Freq: centre, Amp: 0.4},
			{Freq: tuning.FreqAtCents(centre, 15), Amp: 0.4},
		},
	}
	cfg := defaultCfg()
	cfg.ReedCount = 3
	cap := runEngine(t, spec, cfg)

	cap.mu.Lock()
	defer cap.mu.Unlock()
	for i, m := range cap.all {
		if m.ReedsFromBeat {
			t.Fatalf("tick %d: reported %d reeds as a pair, and there are three: %+v",
				i, len(m.Reeds), m.Reeds)
		}
	}
}

// One reed on hand-worked bellows: the stroke modulates its loudness AND its pitch on the same stroke,
// reinforcing one sideband and cancelling the other, so it arrives as a carrier and one companion - a
// pair, to anything reading only the spectrum. The pitch half puts no swing in the amplitude, so it
// does not swing as hard as two lines of that ratio would (see pairDepthSlack).
func TestEngineBellowsModulatedReedIsNotAPair(t *testing.T) {
	const (
		sr      = 48000
		f0      = 262.9 // the C '8 recording
		amIndex = 0.30  // how far the bellows swing its loudness
		bellows = 0.67  // and how fast
	)
	for _, dev := range []float64{0.3, 0.6, 0.9} {
		samples := make([]float32, sr*10)
		phase := 0.0
		for i := range samples {
			ts := float64(i) / sr
			f := f0 + dev*math.Sin(2*math.Pi*bellows*ts)
			phase += 2 * math.Pi * f / sr
			amp := 0.4 * (1 + amIndex*math.Sin(2*math.Pi*bellows*ts))
			samples[i] = float32(amp * math.Sin(phase))
		}

		cfg := defaultCfg()
		cfg.ReedCount = 2
		cap := &capture{}
		if err := NewEngine(cfg, cap.emit).Run(context.Background(),
			&sliceSource{samples: samples, rate: sr}); err != nil {
			t.Fatal(err)
		}
		for i, m := range cap.all {
			if m.ReedsFromBeat {
				t.Fatalf("pitch swing %.1f Hz, tick %d: one reed split into %d: %+v",
					dev, i, len(m.Reeds), m.Reeds)
			}
		}
	}
}

// PairDepth is the whole of the last guard, so its shape is pinned: two reeds cannot swing their own
// sum by more than two thirds, and the swing rises with the ratio between them all the way there.
func TestPairDepthIsBoundedAndRises(t *testing.T) {
	if d := PairDepth(1); math.Abs(d-2.0/3.0) > 1e-3 {
		t.Fatalf("two equal reeds swing %.4f, want 2/3", d)
	}
	prev := -1.0
	for r := 0.0; r <= 1.0001; r += 0.05 {
		d := PairDepth(r)
		if d < prev {
			t.Fatalf("depth fell from %.4f to %.4f at r=%.2f", prev, d, r)
		}
		if d > 2.0/3.0+1e-3 {
			t.Fatalf("r=%.2f swings %.4f, past the 2/3 a pair cannot exceed", r, d)
		}
		prev = d
	}
	// A ratio the other way up describes the same pair.
	if math.Abs(PairDepth(0.5)-PairDepth(2)) > 1e-9 {
		t.Fatal("PairDepth must not care which reed is named first")
	}
}
