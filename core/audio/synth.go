// core/audio/synth.go
package audio

import (
	"context"
	"math"
	"math/rand"
	"time"
)

type ReedSpec struct {
	Freq      float64
	Amp       float64
	Harmonics []float64
}

type SynthSpec struct {
	SampleRate int
	Duration   time.Duration
	Reeds      []ReedSpec
	NoiseAmp   float64
	HumFreq    float64
	HumAmp     float64
	Seed       int64
	Realtime   bool
}

type SynthSource struct {
	spec   SynthSpec
	cancel context.CancelFunc
}

func NewSynthSource(spec SynthSpec) *SynthSource {
	if spec.SampleRate == 0 {
		spec.SampleRate = 48000
	}
	return &SynthSource{spec: spec}
}

func (s *SynthSource) Info() SourceInfo {
	return SourceInfo{Name: "synth", SampleRate: s.spec.SampleRate, Realtime: s.spec.Realtime}
}

func (s *SynthSource) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

func (s *SynthSource) Start(ctx context.Context) (<-chan Block, error) {
	ctx, s.cancel = context.WithCancel(ctx)
	ch := make(chan Block, 4)
	go s.run(ctx, ch)
	return ch, nil
}

func (s *SynthSource) run(ctx context.Context, ch chan<- Block) {
	defer close(ch)
	sr := float64(s.spec.SampleRate)
	total := int(s.spec.Duration.Seconds() * sr)
	rng := rand.New(rand.NewSource(s.spec.Seed))

	// one phase accumulator per partial keeps long signals coherent
	type osc struct{ freq, amp, phase float64 }
	var oscs []osc
	for _, r := range s.spec.Reeds {
		oscs = append(oscs, osc{freq: r.Freq, amp: r.Amp})
		for h, a := range r.Harmonics {
			oscs = append(oscs, osc{freq: r.Freq * float64(h+2), amp: r.Amp * a})
		}
	}
	if s.spec.HumFreq > 0 {
		oscs = append(oscs, osc{freq: s.spec.HumFreq, amp: s.spec.HumAmp})
	}

	blockDur := time.Duration(float64(blockSize) / sr * float64(time.Second))
	for produced := 0; produced < total; {
		n := blockSize
		if total-produced < n {
			n = total - produced
		}
		buf := make([]float32, n)
		for i := 0; i < n; i++ {
			var v float64
			for j := range oscs {
				v += oscs[j].amp * math.Sin(oscs[j].phase)
				oscs[j].phase += 2 * math.Pi * oscs[j].freq / sr
				if oscs[j].phase > 2*math.Pi {
					oscs[j].phase -= 2 * math.Pi
				}
			}
			if s.spec.NoiseAmp > 0 {
				v += s.spec.NoiseAmp * (2*rng.Float64() - 1)
			}
			buf[i] = float32(v)
		}
		select {
		case <-ctx.Done():
			return
		case ch <- Block{Samples: buf, SampleRate: s.spec.SampleRate, Time: time.Now()}:
		}
		produced += n
		if s.spec.Realtime {
			select {
			case <-ctx.Done():
				return
			case <-time.After(blockDur):
			}
		}
	}
}
