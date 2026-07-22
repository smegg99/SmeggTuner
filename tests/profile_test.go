package tests

// The calibration payoff, end to end on real audio: profile the 16' rank from its solo recording,
// then mix that recording with the same key's solo 8' - a true two-voice signal with a known
// answer - and the profiled engine must confirm both voices.

import (
	"context"
	"sort"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/audio"
	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/tuning"
)

// samplesOf decodes a recording to raw samples.
func samplesOf(t *testing.T, path string) ([]float32, int) {
	t.Helper()
	src, err := audio.NewFileSource(path, false, false)
	if err != nil {
		t.Skipf("sample not present: %v", err)
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
	cancel  context.CancelFunc
}

func newMixSource(a, b []float32, rate int) *mixSource {
	n := min(len(a), len(b))
	sum := make([]float32, n)
	for i := range sum {
		sum[i] = a[i] + b[i]
	}
	return &mixSource{samples: sum, rate: rate}
}

func (s *mixSource) Info() audio.SourceInfo {
	return audio.SourceInfo{Name: "mix", SampleRate: s.rate}
}

func (s *mixSource) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

func (s *mixSource) Start(ctx context.Context) (<-chan audio.Block, error) {
	ctx, s.cancel = context.WithCancel(ctx)
	ch := make(chan audio.Block, 4)
	go func() {
		defer close(ch)
		const block = 4096
		for at := 0; at < len(s.samples); at += block {
			end := min(at+block, len(s.samples))
			buf := make([]float32, end-at)
			copy(buf, s.samples[at:end])
			select {
			case <-ctx.Done():
				return
			case ch <- audio.Block{Samples: buf, SampleRate: s.rate}:
			}
		}
	}()
	return ch, nil
}

func TestProfileConfirmsBothVoicesOnRealAudio(t *testing.T) {
	if testing.Short() {
		t.Skip("loads multi-megabyte recordings")
	}

	// Calibrate: the 16' alone, profiling on - what the sweep of a solo register does.
	lone, rate := samplesOf(t, "../sounds/A '16.wav")
	var ratios []float64
	eng := dsp.NewEngine(dsp.EngineConfig{
		A4: 440, ReedCount: 1, FineWindow: 3 * time.Second,
		Octaves:          []dsp.OctaveRequest{{Offset: -12, Reeds: 1}},
		ProfileHarmonics: true,
	}, func(m dsp.Measurement) {
		if len(m.Reeds) == 1 && len(m.Reeds[0].Harmonics) > 0 {
			ratios = append(ratios, m.Reeds[0].Harmonics[0])
		}
	})
	src, err := audio.NewFileSource("../sounds/A '16.wav", false, false)
	if err != nil {
		t.Skip("sample not present")
	}
	if err := eng.Run(context.Background(), src); err != nil {
		t.Fatal(err)
	}
	if len(ratios) == 0 {
		t.Fatal("profiling produced no partial ratios")
	}
	sort.Float64s(ratios)
	r2 := ratios[len(ratios)/2]
	if r2 < 0.05 || r2 > 0.6 {
		t.Fatalf("calibrated H2/H1 %.3f outside anything a reed sounds", r2)
	}

	// Tune: both solo takes mixed - the 16' and the genuine 8' of the same key, 1.0 Hz apart.
	eight, _ := samplesOf(t, "../sounds/A '8.wav")
	var all []dsp.Measurement
	eng = dsp.NewEngine(dsp.EngineConfig{
		A4: 440, ReedCount: 2, FineWindow: 3 * time.Second,
		Octaves:  []dsp.OctaveRequest{{Offset: -12, Reeds: 1}, {Offset: 0, Reeds: 1}},
		Profiles: []dsp.RankProfile{{Offset: -12, Note: tuning.Note(57), R2: r2}},
	}, func(m dsp.Measurement) { all = append(all, m) })
	if err := eng.Run(context.Background(), newMixSource(lone, eight, rate)); err != nil {
		t.Fatal(err)
	}
	m, ok := dsp.Aggregate(all, 440, 0.5)
	if !ok {
		t.Fatal("no usable measurements")
	}

	if m.Note != tuning.Note(69) {
		t.Errorf("note %s, want the key A4", m.Note.Name(tuning.NamingCDEFGAB))
	}
	if len(m.Reeds) != 2 {
		t.Fatalf("both ranks sound and the profile knows it, got %+v", m.Reeds)
	}
	if len(m.Bands) != 2 || m.Bands[1].GhostOnly {
		t.Errorf("the 8' is a confirmed voice, got %+v", m.Bands)
	}
}
