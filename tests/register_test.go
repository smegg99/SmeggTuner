package tests

import (
	"context"
	"errors"
	"math"
	"os"
	"testing"

	"smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/dsp"
)

// A lone rank under a 16'+8' register lands in the lowest slot, reading exactly what the plain tuner reads.
func TestALoneRankUnderATwoOctaveRegisterFillsTheLowestSlot(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("loads two dozen recordings")
	}
	for _, s := range loneRanks(t) {
		t.Run(s.name, func(t *testing.T) {
			want := loneRankWant(s)
			all := runFile(t, s.path, registerConfig(t, loneRankA4, "LM")).all
			key := want + 12
			skipStray(t, s, reads(all, key), all)
			if !reads(all, key) {
				t.Fatalf("key reads %s, want %s", summary(all), name(key))
			}
			got, n := sustained(sounding(all, key), 0)
			plain, _ := sustained(sounding(loneRun(t, s).all, want), 0)
			if n == 0 || got != plain {
				t.Errorf("the 16' slot reads %.4f Hz over %d analyses, the plain tuner %.4f", got, n, plain)
			}
			for _, m := range measured(all, key) {
				if m.Reeds[0].Octave != -12 {
					t.Fatalf("the lowest slot at octave %d", m.Reeds[0].Octave)
				}
			}
		})
	}
}

func samplesOf(t *testing.T, path string) ([]float32, int) {
	t.Helper()
	src, err := audio.NewFileSource(path, false, false)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			t.Skipf("optional recording not present: %v", err)
		}
		t.Fatalf("recording failed: %v", err)
	}
	blocks, err := src.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	var all []float32
	for b := range blocks {
		all = append(all, b.Samples...)
	}
	_ = src.Stop()
	return all, src.Info().SampleRate
}

// mixSource plays the sum of two recordings, as if both ranks sounded at once.
type mixSource struct {
	samples []float32
	rate    int
}

func (s mixSource) Info() audio.SourceInfo { return audio.SourceInfo{Name: "mix", SampleRate: s.rate} }
func (s mixSource) Stop() error            { return nil }

func (s mixSource) Start(ctx context.Context) (<-chan audio.Block, error) {
	ch := make(chan audio.Block)
	go func() {
		defer close(ch)
		for at := 0; at < len(s.samples); at += 4096 {
			b := append([]float32(nil), s.samples[at:min(at+4096, len(s.samples))]...)
			select {
			case ch <- audio.Block{Samples: b, SampleRate: s.rate}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

// The same key's 16' and 8' recorded alone, summed: each slot must read what its rank reads alone.
func TestAMixedSixteenAndEightReadsEachRankInItsSlot(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("loads recordings")
	}
	const low, high = "../sounds/A '16.wav", "../sounds/A '8.wav"
	a, rate := samplesOf(t, low)
	b, _ := samplesOf(t, high)
	sum := make([]float32, min(len(a), len(b)))
	for i := range sum {
		sum[i] = a[i] + b[i]
	}
	all := runSource(t, registerConfig(t, loneRankA4, "LM"), mixSource{sum, rate}).all
	if !reads(all, 69) {
		t.Fatalf("reads %s, want the key A4", summary(all))
	}
	frames := sounding(all, 69)
	alone16, _ := sustained(sounding(runFile(t, low, dsp.EngineConfig{A4: loneRankA4, ReedCount: 1}).all, 57), 0)
	alone8, _ := sustained(sounding(runFile(t, high, dsp.EngineConfig{A4: loneRankA4, ReedCount: 1}).all, 69), 0)
	got16, n16 := sustained(frames, 0)
	got8, n8 := sustained(frames, 1)
	t.Logf("16' %.3f Hz mixed, %.3f alone; 8' %.3f Hz mixed, %.3f alone", got16, alone16, got8, alone8)
	if n16 == 0 || math.Abs(got16-alone16) > 0.1 {
		t.Errorf("the 16' slot at %.3f Hz over %d of %d analyses, the rank alone reads %.3f", got16, n16, len(frames), alone16)
	}
	// The 8' shares its band with the 16's second partial, 0.7 Hz above it here.
	if n8 == 0 || math.Abs(got8-alone8) > 0.1 {
		t.Errorf("the 8' slot at %.3f Hz over %d of %d analyses, the rank alone reads %.3f", got8, n8, len(frames), alone8)
	}
}
