package dsp

import (
	"context"
	"math"
	"testing"

	"smegg.me/smeggtuner/backend/core/audio"
)

// waveSource plays wave(i) for seconds in 4096-frame blocks; lostAt marks blocks after a queue drop.
type waveSource struct {
	seconds float64
	wave    func(i int) float32
	lostAt  map[int]int
}

func (s waveSource) Info() audio.SourceInfo { return audio.SourceInfo{Name: "wave", SampleRate: 48000} }
func (s waveSource) Stop() error            { return nil }
func (s waveSource) Start(ctx context.Context) (<-chan audio.Block, error) {
	ch := make(chan audio.Block)
	go func() {
		defer close(ch)
		total := int(s.seconds * 48000)
		for at, n := 0, 0; at < total; at, n = at+4096, n+1 {
			b := make([]float32, min(4096, total-at))
			for i := range b {
				b[i] = s.wave(at + i)
			}
			select {
			case ch <- audio.Block{Samples: b, SampleRate: 48000, Lost: s.lostAt[n]}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return ch, nil
}

func runWave(t *testing.T, s waveSource) []Measurement {
	t.Helper()
	var out []Measurement
	if err := NewEngine(EngineConfig{}, func(m Measurement) { out = append(out, m) }).Run(context.Background(), s); err != nil {
		t.Fatal(err)
	}
	return out
}

func sine(amp float64) func(int) float32 {
	return func(i int) float32 { return float32(amp * math.Sin(2*math.Pi*440.5*float64(i)/48000)) }
}

func TestEngineInputIntervalsCountHalfSeconds(t *testing.T) {
	ms := runWave(t, waveSource{seconds: 4, wave: sine(0.2)})
	var seen []InputInterval
	for _, m := range ms {
		if n := len(seen); m.Input.Seq != 0 && (n == 0 || seen[n-1].Seq != m.Input.Seq) {
			seen = append(seen, m.Input)
		}
	}
	if len(seen) < 7 {
		t.Fatalf("%d intervals over four seconds", len(seen))
	}
	for i, iv := range seen {
		if iv.Seq != uint64(i+1) || iv.Restart != (i == 0) || iv.Clipped {
			t.Fatalf("interval %d: %+v", i, iv)
		}
		if want := 0.2 * 32767 / math.Sqrt2; math.Abs(float64(iv.RMS)-want) > 30 || iv.Peak < 6500 || iv.Peak > 6554 {
			t.Fatalf("interval %d level %+v want rms %.0f", i, iv, want)
		}
	}
}

// A flat top is too loud; a full-scale sine that never repeats a sample is not.
func TestEngineTooLoudIsTheFlatTopLatch(t *testing.T) {
	clipped := func(i int) float32 {
		return float32(max(-0.3, min(0.3, 0.6*math.Sin(2*math.Pi*440.5*float64(i)/48000))))
	}
	loud := 0
	for _, m := range runWave(t, waveSource{seconds: 3, wave: clipped}) {
		if m.State == StateTooLoud {
			loud++
		}
		if m.Input.Seq > 0 && !m.Input.Clipped {
			t.Fatalf("interval %+v", m.Input)
		}
	}
	if loud == 0 {
		t.Fatal("a flat-topped capture never showed too loud")
	}
	for _, m := range runWave(t, waveSource{seconds: 3, wave: sine(0.99)}) {
		if m.State == StateTooLoud || m.Input.Clipped {
			t.Fatalf("a clean sine is not clipped: %+v", m.Input)
		}
	}
}

func TestEngineUnreliableHoldsATenthOfASecondAfterALoss(t *testing.T) {
	ms := runWave(t, waveSource{seconds: 3, wave: sine(0.2), lostAt: map[int]int{20: 512}})
	flagged := 0
	for _, m := range ms {
		if m.Unreliable {
			flagged++
		}
	}
	// The lossy block and 0.1 s after it span two more 4096-frame blocks: about three heartbeats.
	if flagged == 0 || flagged > 6 {
		t.Fatalf("%d measurements flagged unreliable", flagged)
	}
	if ms[len(ms)-1].Unreliable {
		t.Fatal("the flag must clear")
	}
}
