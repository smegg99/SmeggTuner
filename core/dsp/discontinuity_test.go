package dsp

import (
	"context"
	"math"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/audio"
)

// jumpSource replays a fixed list of blocks, At and all, so a test can put a SEEK in the stream. It is
// a Transport (has a timeline), which turns the engine's discontinuity handling on.
type jumpSource struct {
	blocks []audio.Block
	rate   int
}

func (s *jumpSource) Info() audio.SourceInfo {
	return audio.SourceInfo{Name: "jump", SampleRate: s.rate}
}

func (s *jumpSource) Stop() error { return nil }

func (s *jumpSource) Start(ctx context.Context) (<-chan audio.Block, error) {
	ch := make(chan audio.Block, 4)
	go func() {
		defer close(ch)
		for _, b := range s.blocks {
			select {
			case <-ctx.Done():
				return
			case ch <- b:
			}
		}
	}()
	return ch, nil
}

// audio.Transport, stubbed: the engine reads only Block.At off the stream, so none of these are exercised.
func (s *jumpSource) Duration() time.Duration                   { return 0 }
func (s *jumpSource) Position() time.Duration                   { return 0 }
func (s *jumpSource) Selection() (time.Duration, time.Duration) { return 0, 0 }
func (s *jumpSource) Paused() bool                              { return false }
func (s *jumpSource) Moving() bool                              { return false }
func (s *jumpSource) Seek(time.Duration)                        {}
func (s *jumpSource) SetPaused(bool)                            {}
func (s *jumpSource) SetRange(time.Duration, time.Duration)     {}
func (s *jumpSource) SetLoop(bool)                              {}
func (s *jumpSource) Peaks(time.Duration, time.Duration, int) []audio.Peak {
	return nil
}

// toneBlocks slices dur of a single continuous sine into file-sized blocks, At starting atStart
// samples in. The phase runs unbroken, so the only seam is the one a test puts there deliberately.
func toneBlocks(freq, amp float64, dur time.Duration, atStart, rate int) []audio.Block {
	n := int(dur.Seconds() * float64(rate))
	const block = 1024
	var out []audio.Block
	for off := 0; off < n; off += block {
		end := off + block
		if end > n {
			end = n
		}
		buf := make([]float32, end-off)
		for i := range buf {
			t := float64(off+i) / float64(rate)
			buf[i] = float32(amp * math.Sin(2*math.Pi*freq*t))
		}
		start := atStart + off
		out = append(out, audio.Block{
			Samples:    buf,
			SampleRate: rate,
			Time:       time.Now(),
			At:         time.Duration(float64(start) / float64(rate) * float64(time.Second)),
		})
	}
	return out
}

// TestEngineResetsAcrossASeek is the pause/stop bug made small: one A4 reed sounds throughout, but the
// transport is MOVED mid-stream, and the reed's pitch there differs. An engine that analyses across
// the seam holds two pitches in one window and reports a beat between reeds never sounding together.
func TestEngineResetsAcrossASeek(t *testing.T) {
	const rate = 48000

	blocks := toneBlocks(442.6, 0.4, 4*time.Second, 0, rate)
	blocks = append(blocks, toneBlocks(445.0, 0.4, 2*time.Second, 20*rate, rate)...)

	cap := &capture{}
	e := NewEngine(EngineConfig{
		A4: 440, ReedCount: 2, FineWindow: 3 * time.Second, Highpass: true,
	}, cap.emit)
	if err := e.Run(context.Background(), &jumpSource{blocks: blocks, rate: rate}); err != nil {
		t.Fatal(err)
	}

	// The corruption is a transient (it lives only while the window straddles the seam), so the whole
	// stream is checked, not just the last reading.
	for i, m := range cap.all {
		if len(m.Beats) != 0 {
			t.Fatalf("measurement %d: a seek was read as a beat between reeds that never sounded together: reeds=%+v beats=%+v", i, m.Reeds, m.Beats)
		}
		if m.ReedsSeparated && len(m.Reeds) > 1 {
			t.Fatalf("measurement %d: a seek was read as a second reed: reeds=%+v", i, m.Reeds)
		}
	}
}
