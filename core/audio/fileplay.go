// core/audio/fileplay.go
package audio

import (
	"context"
	"time"
)

// deliverPoll is how often the producer checks what the speaker has reached while pacing.
const deliverPoll = 4 * time.Millisecond

// maxLeadIn is how far ahead of real time the source may run while the speaker fills its
// cushion, so the cushion fills instantly. Bounded: a machine with no output never plays.
const maxLeadIn = 2 * time.Second

// maxHold is the longest a block may wait for the speaker before going to the engine anyway -
// the fallback for a machine with no working output, whose played-count never moves.
const maxHold = time.Second

// Start begins playback. Unlike the mic, a FileSource may be started again after it has been
// stopped, keeping its playhead, selection and loop flag - which is what lets the transport
// the technician steered outlive the engine that measures it.
func (s *FileSource) Start(ctx context.Context) (<-chan Block, error) {
	// Cancelled AND WAITED FOR: a cancelled goroutine is not a dead one, and two producers
	// writing the one lock-free speaker ring corrupts it.
	if s.cancel != nil {
		s.cancel()
		<-s.done
	}
	s.done = make(chan struct{})

	// The last run's queued audio was never heard: the playhead takes it back and the speaker
	// starts empty (full cushion). See dropQueue and Speaker.Reset.
	s.dropQueue()

	s.mu.Lock()
	if s.parked || s.pos >= s.lastPlayable() {
		s.pos = s.from
	}
	s.parked = false
	s.mu.Unlock()

	ctx, s.cancel = context.WithCancel(ctx)
	done := s.done
	// Whatever sits here is audio the engine has but the speaker has not played: how far the
	// reading leads the sound in the room. Deep enough to fill the cushion once, no deeper.
	ch := make(chan Block, 6)
	go func() {
		defer close(done)
		defer close(ch)

		// Stopped by the user: the playhead keeps only what was heard. A file that simply
		// ENDED is not stopped and its tail is left to play out.
		defer func() {
			if ctx.Err() != nil {
				s.dropQueue()
			}
		}()

		blockDur := time.Duration(float64(fileBlockSize) / float64(s.sampleRate) * float64(time.Second))

		// The clock is absolute: wake on a deadline advancing by one block rather than
		// sleeping one block, or the source drifts slower than real time and the card underruns.
		next := time.Now()

		// held blocks wait until the speaker has played them, then go to the engine.
		var held []waiting
		gen := s.generation()

		for {
			sink := s.currentSink()

			// The queue was thrown away (pause/stop/start): the playhead was rewound over
			// these blocks, about to be produced again.
			if now := s.generation(); now != gen {
				gen = now
				held = held[:0]
			}

			block, at, done := s.nextBlock()
			if done {
				return
			}

			if block == nil {
				// Paused or parked. Whatever the speaker still has to play is still owed to
				// the engine, so keep feeding it while we idle.
				if s.deliverHeard(ctx, ch, sink, &held) {
					return
				}

				select {
				case <-ctx.Done():
					return
				case <-time.After(blockDur):
				}

				next = time.Now()
				continue
			}

			if sink == nil || !s.realtime {
				// Nobody listening: the engine takes the audio at once (the headless path).
				select {
				case <-ctx.Done():
					return
				case ch <- Block{
					Samples:    block,
					SampleRate: s.sampleRate,
					Time:       time.Now(),
					At:         s.timeAt(at),
				}:
				}
				continue
			}

			// Heard before it is measured, out of the same slice: the engine filters in
			// place, so the speaker is given the block first. Sink.Write copies.
			playedEnd := s.markBlock(at, len(block))
			sink.Write(block, s.sampleRate)

			held = append(held, waiting{
				block:     block,
				at:        at,
				playedEnd: playedEnd,
				since:     time.Now(),
			})

			if s.deliverHeard(ctx, ch, sink, &held) {
				return
			}

			if !s.realtime {
				continue
			}

			next = next.Add(blockDur)

			// The machine stalled past what the buffer can hide. Do not sprint to catch up:
			// give up the lost time and carry on from now.
			if late := time.Since(next); late > time.Second {
				next = time.Now()
				continue
			}

			wait := time.Until(next)

			// The speaker has not started yet, so do not sleep: fill it. The lead is real
			// time we have chosen not to wait out, so it is NOT carried forward - carrying it
			// made the card drain the whole cushion the instant it started.
			if wait > 0 && wait < maxLeadIn && sink != nil && !sink.Playing() {
				next = time.Now()
				continue
			}

			// The wait is spent delivering, not slept through: the speaker reaches blocks
			// while this goroutine sleeps, so wake often and hand over what it reached.
			for wait > 0 {
				nap := wait
				if nap > deliverPoll {
					nap = deliverPoll
				}

				select {
				case <-ctx.Done():
					return
				case <-time.After(nap):
				}

				if s.deliverHeard(ctx, ch, sink, &held) {
					return
				}
				wait = time.Until(next)
			}
		}
	}()
	return ch, nil
}
