//go:build !js

// backend/core/audio/ring_test.go
package audio

import (
	"sync"
	"testing"
)

func TestRingOverflowPreservesUnreadSamples(t *testing.T) {
	sp := loaded(t, 8)
	sp.enqueue([]float32{1, 2, 3})
	for _, v := range pullFrames(sp, 8) {
		if v != 0.5 {
			t.Fatalf("overflow overwrote unread sample: %v", v)
		}
	}
	if _, drops := sp.Stats(); drops != 3 {
		t.Fatalf("drops=%d", drops)
	}
	sp.enqueue([]float32{1, 2, 3})
	got := pullFrames(sp, 3)
	for i, v := range got {
		if v != float32(i+1) {
			t.Fatalf("wrapped sample %d = %v", i, v)
		}
	}
}

func TestRingProducerAndCallbackDoNotShareUnreadCells(t *testing.T) {
	sp := loaded(t, 64)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		out := make([]byte, 17*4)
		for range 10000 {
			sp.pull(out, nil, 17)
		}
	}()
	samples := make([]float32, 31)
	for i := range samples {
		samples[i] = 0.25
	}
	for range 10000 {
		sp.enqueue(samples)
	}
	wg.Wait()
	if got := sp.Buffered(); got < 0 || got > 64 {
		t.Fatalf("buffered=%d", got)
	}
}
