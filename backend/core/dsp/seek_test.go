package dsp

import (
	"context"
	"math"
	"testing"
	"time"

	"smegg.me/smeggtuner/backend/core/audio"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// seekSource plays A4 from 0 s, then jumps the playhead to 20 s and plays C5: a recording seeked mid-run.
type seekSource struct {
	audio.Transport // only marks the source as a timeline; Run reads Block.At, not these methods
}

func (seekSource) Info() audio.SourceInfo { return audio.SourceInfo{Name: "seek", SampleRate: 48000} }
func (seekSource) Stop() error            { return nil }
func (seekSource) Start(ctx context.Context) (<-chan audio.Block, error) {
	ch := make(chan audio.Block)
	go func() {
		defer close(ch)
		play := func(from, seconds, freq float64) bool {
			total := int(seconds * 48000)
			for at := 0; at < total; at += 4096 {
				b := make([]float32, min(4096, total-at))
				for i := range b {
					b[i] = float32(0.2 * math.Sin(2*math.Pi*freq*float64(at+i)/48000))
				}
				offset := time.Duration((from + float64(at)/48000) * float64(time.Second))
				select {
				case ch <- audio.Block{Samples: b, SampleRate: 48000, At: offset}:
				case <-ctx.Done():
					return false
				}
			}
			return true
		}
		if play(0, 6, 440) {
			play(20, 6, 523.25)
		}
	}()
	return ch, nil
}

// After a seek nothing from before it survives: every reading is the new note, and the lock is earned anew.
func TestEngineSeekStartsTheAnalysisAfresh(t *testing.T) {
	var ms []Measurement
	e := NewEngine(EngineConfig{}, func(m Measurement) { ms = append(ms, m) })
	if err := e.Run(context.Background(), seekSource{}); err != nil {
		t.Fatal(err)
	}
	measured := 0
	for _, m := range ms {
		if m.SourceAt < 20 {
			continue
		}
		if m.Lock.Visible && measured < 7 {
			t.Fatalf("a lock shows %.2f s after the seek, after %d analyses of the new note, not eight", m.SourceAt-20, measured+1)
		}
		if m.Frame == FrameMeasured {
			measured++
			if m.Note != tuning.Note(72) {
				t.Fatalf("read %s %.2f s after the seek, want C5", m.NoteName, m.SourceAt-20)
			}
		}
	}
	if measured == 0 {
		t.Fatal("nothing measured after the seek")
	}
}
