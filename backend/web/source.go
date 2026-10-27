// backend/web/source.go
package web

import (
	"context"
	"sync"

	"smegg.me/smeggtuner/backend/core/audio"
)

// queueBlocks bounds the blocks waiting for the engine; older ones are dropped so readings never trail the sound.
const queueBlocks = 8

// BrowserSource is the microphone as the worker hands it over, block by block.
type BrowserSource struct {
	rate int

	mu   sync.Mutex
	ch   chan audio.Block
	lost int // frames dropped here, charged to the next block
}

// NewBrowserSource is a source at the AudioContext's sample rate.
func NewBrowserSource(rate int) *BrowserSource { return &BrowserSource{rate: rate} }

func (s *BrowserSource) Start(context.Context) (<-chan audio.Block, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ch = make(chan audio.Block, queueBlocks)
	s.lost = 0
	return s.ch, nil
}

func (s *BrowserSource) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ch != nil {
		close(s.ch)
		s.ch = nil
	}
	return nil
}

func (s *BrowserSource) Info() audio.SourceInfo {
	return audio.SourceInfo{Name: "browser", SampleRate: s.rate, Realtime: true}
}

// Push hands one block to the engine without blocking; lost is what the worker already dropped before it.
func (s *BrowserSource) Push(samples []float32, lost int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.ch == nil {
		return
	}
	b := audio.Block{Samples: samples, SampleRate: s.rate, Lost: lost + s.lost}
	select {
	case s.ch <- b:
		s.lost = 0
		return
	default:
	}
	select {
	case old := <-s.ch:
		b.Lost += len(old.Samples) + old.Lost
	default:
	}
	select {
	case s.ch <- b:
		s.lost = 0
	default:
		s.lost = b.Lost + len(samples)
	}
}
