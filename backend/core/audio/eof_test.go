//go:build !js

// backend/core/audio/eof_test.go
package audio

import (
	"context"
	"testing"
	"time"
)

func shortSource() *FileSource {
	samples := make([]float32, 4096)
	for i := range samples {
		samples[i] = float32(i)
	}
	return &FileSource{samples: samples, sampleRate: 48000, realtime: true, to: len(samples)}
}

func TestFileEOFDrainsHeldAnalysisBlocks(t *testing.T) {
	s := shortSource()
	s.SetSink(&drainingSink{})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	blocks, err := s.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	total := 0
	for b := range blocks {
		for _, v := range b.Samples {
			if v != float32(total) {
				t.Fatalf("sample %d = %v", total, v)
			}
			total++
		}
	}
	if ctx.Err() != nil {
		t.Fatal("EOF did not close after draining")
	}
	if total != len(s.samples) {
		t.Fatalf("EOF lost held samples: got %d want %d", total, len(s.samples))
	}
}

func TestFileRealtimeWithoutSinkStillPaces(t *testing.T) {
	s := shortSource()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	start := time.Now()
	blocks, err := s.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	total := 0
	for b := range blocks {
		total += len(b.Samples)
	}
	if total != len(s.samples) {
		t.Fatalf("samples=%d", total)
	}
	if elapsed := time.Since(start); elapsed < 75*time.Millisecond {
		t.Fatalf("realtime source decoded unpaced in %v", elapsed)
	}
}

func TestFileEOFDrainStopsOnCancellation(t *testing.T) {
	s := shortSource()
	s.SetSink(&drainingSink{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	blocks, err := s.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	<-blocks
	cancel()
	done := make(chan struct{})
	go func() {
		for range blocks {
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("EOF drain ignored cancellation")
	}
}

type partialSink struct {
	drainingSink
	flushed bool
}

func (s *partialSink) Flush()        { s.flushed = true; s.play(s.Buffered()) }
func (s *partialSink) Playing() bool { return false }

func TestShortFileStartsPartialSpeakerCushion(t *testing.T) {
	s := shortSource()
	sink := &partialSink{}
	s.SetSink(sink)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	blocks, err := s.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	total := 0
	for b := range blocks {
		total += len(b.Samples)
	}
	if !sink.flushed || sink.Played() != len(s.samples) || total != len(s.samples) {
		t.Fatalf("partial cushion: flushed=%v played=%d analyzed=%d", sink.flushed, sink.Played(), total)
	}
	if ctx.Err() != nil {
		t.Fatal("short file timed out")
	}
}

func TestPauseAfterLoopWrapUsesHeardPosition(t *testing.T) {
	s := shortSource()
	s.loop = true
	s.from = 1024
	s.to = 3072
	s.pos = 1024
	sink := &drainingSink{}
	s.SetSink(sink)
	for range 3 {
		block, at, _ := s.nextBlock()
		s.markBlock(at, len(block))
		sink.Write(block, 48000)
	}
	sink.play(1536)
	heard := s.Position()
	s.SetPaused(true)
	if got := s.Position(); got != heard {
		t.Fatalf("pause rewound loop: got %v want %v", got, heard)
	}
}
