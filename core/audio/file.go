// core/audio/file.go
package audio

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-audio/wav"
)

var errUnsupportedFormat = errors.New("unsupported wav format")

// FileSource plays back a WAV file as a stream of Blocks. The whole file is decoded at
// open time and downmixed to mono. With realtime=true emission is paced to the file's
// sample rate; with loop=true playback restarts until the context is cancelled.
//
// It is also a Transport: it can be paused, seeked, and confined to a range. The range is
// enforced HERE, in the producer, not in the view - so the engine is never handed a sample
// from outside the selection and the drawn playhead can never point at one.
type FileSource struct {
	name       string
	samples    []float32
	sampleRate int
	realtime   bool
	cancel     context.CancelFunc
	// closed once the playback goroutine has actually exited, so the next Start knows it
	// is the only producer. See Start.
	done chan struct{}

	// Where the audio is heard, or nil for a run nobody is listening to.
	sink Sink

	// The needle's timeline. The speaker only counts samples played and cannot say which
	// part of the recording they came from (and after a loop the answer is not increasing),
	// so every handed-over block is recorded here - where it starts in the FILE and in the
	// speaker's own count - and Position is a lookup, never a subtraction.
	marks   []mark
	written int // samples handed to the speaker since it was last reset

	// Everything the transport can change while the goroutine runs. pos, from and to are
	// sample offsets; from inclusive, to exclusive.
	mu     sync.Mutex
	pos    int
	from   int
	to     int
	paused bool
	loop   bool
	// parked: the selection has played out and there is nothing left to hand over. Cleared
	// by anything that gives the playhead audio to play again.
	parked bool
	// gen counts the times the queue has been thrown away. See generation().
	gen int
}

func NewFileSource(path string, realtime, loop bool) (*FileSource, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	dec := wav.NewDecoder(f)
	buf, err := dec.FullPCMBuffer()
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	if buf.Format == nil || len(buf.Data) == 0 {
		return nil, fmt.Errorf("decode %s: empty or invalid wav", path)
	}
	mono, err := downmixMono(buf)
	if err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return &FileSource{
		name:       filepath.Base(path),
		samples:    mono,
		sampleRate: buf.Format.SampleRate,
		realtime:   realtime,
		loop:       loop,
		from:       0,
		to:         len(mono),
	}, nil
}

func (s *FileSource) Info() SourceInfo {
	return SourceInfo{Name: s.name, SampleRate: s.sampleRate, Realtime: s.realtime}
}

// SetSink is where the audio is heard; nil plays nothing. Only fed on the realtime path -
// the headless one decodes far faster than real time and must not be piped at a sound card.
func (s *FileSource) SetSink(sink Sink) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sink = sink
}

func (s *FileSource) Stop() error {
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

// nextBlock takes the next run of samples to emit and advances the playhead. It returns nil
// when there is nothing to emit (paused, or parked at the end of a non-looping selection),
// both live states. done reports a non-looping source running off the end of the WHOLE file.
func (s *FileSource) nextBlock() (block []float32, at int, done bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.paused || s.parked {
		return nil, 0, false
	}
	if s.pos < s.from || s.pos >= s.to {
		if !s.loop {
			// parked is a STATE, not a position: the playhead rests on the last sample
			// (where the audio stopped) but the flag is what says nothing is left to play.
			s.parked = true
			s.pos = s.lastPlayable()
			return nil, 0, s.to >= len(s.samples)
		}
		s.pos = s.from
	}

	end := s.pos + fileBlockSize
	if end > s.to {
		end = s.to
	}

	// The consumer owns the samples and the engine filters them in place, so a window into
	// s.samples would let the first pass corrupt the decoded audio for every later loop.
	block = make([]float32, end-s.pos)
	copy(block, s.samples[s.pos:end])

	at = s.pos
	s.pos = end
	return block, at, false
}
