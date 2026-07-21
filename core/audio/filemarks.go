// core/audio/filemarks.go
package audio

import (
	"context"
	"time"
)

// waiting is a block given to the speaker, held until the speaker plays it before it goes to
// the engine. See the producer loop.
type waiting struct {
	block     []float32
	at        int
	playedEnd int // the speaker's played-count at which this block has been fully heard
	since     time.Time
}

// mark ties one handed-over block to the speaker's own count: `played` samples in lies
// `file`, the first sample of this block in the recording.
type mark struct {
	played int // the speaker's count at the first sample of this block
	file   int // where that sample is in the file
	n      int // how many samples the block holds
}

// deliverHeard hands the engine every block the speaker has now played, so the reading and
// the sound describe the same moment. Reports whether the context ended.
func (s *FileSource) deliverHeard(ctx context.Context, ch chan<- Block, sink Sink, held *[]waiting) bool {
	// Nobody listening: nothing is held back. The headless path takes this every time.
	if sink == nil || len(*held) == 0 {
		return false
	}

	played := sink.Played()

	for len(*held) > 0 {
		w := (*held)[0]

		// The block goes to the engine once the speaker has STARTED it, not finished it, so
		// the audio in the room right now is audio the engine holds.
		if played < w.playedEnd-len(w.block) && time.Since(w.since) < maxHold {
			return false
		}
		*held = (*held)[1:]

		select {
		case <-ctx.Done():
			return true
		case ch <- Block{
			Samples:    w.block,
			SampleRate: s.sampleRate,
			Time:       time.Now(),
			At:         s.timeAt(w.at),
		}:
		}
	}
	return false
}

func (s *FileSource) currentSink() Sink {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sink
}

// markBlock records where a block came from against the speaker's running played-count, and
// returns the count at which this block will have finished playing.
func (s *FileSource) markBlock(file, n int) (playedEnd int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.marks = append(s.marks, mark{played: s.written, file: file, n: n})
	s.written += n
	return s.written
}

// generation counts how many times the queue has been thrown away, so held blocks that a
// pause/stop rewound over are discarded instead of re-delivered.
func (s *FileSource) generation() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.gen
}

// forgetMarks drops every mark: the speaker's count has been reset. Caller holds mu.
func (s *FileSource) forgetMarks() {
	s.marks = nil
	s.written = 0
}

// lastPlayable is the final sample the playhead may rest on. Caller holds mu.
func (s *FileSource) lastPlayable() int {
	if s.to > s.from {
		return s.to - 1
	}
	return s.from
}
