// backend/web/source_test.go
package web

import (
	"context"
	"testing"
)

func block(v float32) []float32 {
	b := make([]float32, 1024)
	b[0] = v
	return b
}

func TestPushNeverBlocksAndDropsOldest(t *testing.T) {
	s := NewBrowserSource(48000)
	ch, err := s.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for i := range 20 {
		s.Push(block(float32(i)), 0)
	}
	lost := 0
	for want := 12; want < 20; want++ {
		b := <-ch
		if b.Samples[0] != float32(want) {
			t.Fatalf("got block %v, want %d: the newest %d blocks in order", b.Samples[0], want, queueBlocks)
		}
		lost += b.Lost
	}
	if lost != 12*1024 {
		t.Fatalf("Lost total = %d, want %d dropped frames", lost, 12*1024)
	}
	select {
	case b := <-ch:
		t.Fatalf("unexpected extra block %v", b.Samples[0])
	default:
	}
}

func TestPushCarriesCallerLost(t *testing.T) {
	s := NewBrowserSource(48000)
	ch, _ := s.Start(context.Background())
	s.Push(block(1), 512)
	if b := <-ch; b.Lost != 512 || b.SampleRate != 48000 {
		t.Fatalf("block Lost %d rate %d, want 512 and 48000", b.Lost, b.SampleRate)
	}
}

func TestPushWhileStoppedIsDropped(t *testing.T) {
	s := NewBrowserSource(48000)
	s.Push(block(1), 0)
	ch, _ := s.Start(context.Background())
	_ = s.Stop()
	s.Push(block(2), 0)
	if b, ok := <-ch; ok {
		t.Fatalf("got block %v after Stop, want a closed channel", b.Samples[0])
	}
}
