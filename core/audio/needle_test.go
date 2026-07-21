// core/audio/needle_test.go
package audio

import (
	"testing"
	"time"
)

// The playhead never goes backwards. It used to be handovers MINUS queue - two counters read
// at two instants, so a block handed over between them lurched it back by up to a block. It is
// now a lookup against the speaker's own played count, which one thread moves and only upward.
func TestThePlayheadNeverGoesBackwards(t *testing.T) {
	s := openFile(t)
	sink := &drainingSink{}
	s.SetSink(sink)

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 400; i++ {
			block, at, over := s.nextBlock()
			if over {
				return
			}
			if block == nil {
				continue
			}
			s.markBlock(at, len(block))
			sink.Write(block, 48000)
			time.Sleep(time.Millisecond)
		}
	}()

	go func() {
		for {
			select {
			case <-done:
				return
			default:
			}
			sink.play(480) // 10 ms at a time, like a driver period
			time.Sleep(time.Millisecond)
		}
	}()

	var last time.Duration
	for i := 0; i < 3000; i++ {
		now := s.Position()
		if now < last {
			t.Fatalf("the playhead went backwards: %v then %v (a needle that glitches)", last, now)
		}
		last = now
	}
	<-done
}

// A loop is the one time the playhead may go back, and the old subtraction got even that
// wrong: the file offset wraps while the speaker's count keeps climbing.
func TestThePlayheadFollowsALoopRoundAgain(t *testing.T) {
	s := openFile(t)
	s.SetRange(time.Second, 2*time.Second)
	s.SetLoop(true)

	sink := &drainingSink{}
	s.SetSink(sink)

	// the fragment handed over twice over - counted in SAMPLES, so it does not quietly stop
	// testing anything the day the block size changes
	for handed := 0; handed < 2*48000; {
		block, at, _ := s.nextBlock()
		if block == nil {
			continue
		}
		s.markBlock(at, len(block))
		sink.Write(block, 48000)
		handed += len(block)
	}

	var wrapped bool
	var last time.Duration

	for played := 0; played < 90_000; played += 4096 {
		sink.play(4096)

		now := s.Position()
		if now < time.Second || now > 2*time.Second {
			t.Fatalf("the playhead reached %v, outside the fragment it is looping", now)
		}
		if now < last {
			wrapped = true // it came round again, which is the whole point of a loop
		}
		last = now
	}

	if !wrapped {
		t.Fatal("the playhead never came round: a looping needle must return to the top")
	}
}

// Dropping the queue moves neither the needle nor the audio: the queued-but-unheard audio must
// not move the needle, and its samples are handed BACK to the playhead so resume skips none.
func TestDroppingTheQueueMovesNeitherTheNeedleNorTheAudio(t *testing.T) {
	s := openFile(t)
	sink := &drainingSink{}
	s.SetSink(sink)

	// hand over four seconds, of which the card has played three and a half
	for handed := 0; handed < 4*48000; {
		block, at, _ := s.nextBlock()
		if block == nil {
			break
		}
		s.markBlock(at, len(block))
		sink.Write(block, 48000)
		handed += len(block)
	}
	sink.play(175_200) // 3.65 s at 48 kHz

	heard := s.Position()
	if heard < 3*time.Second || heard > 4*time.Second {
		t.Fatalf("the needle is at %v, which is not where this test put it", heard)
	}

	s.SetPaused(true)

	if got := s.Position(); got != heard {
		t.Fatalf("the needle jumped from %v to %v when the queue was dropped", heard, got)
	}

	// and playback resumes from the sound that was actually heard, not the audio thrown away
	block, at, _ := s.nextBlock()
	s.SetPaused(false)
	block, at, _ = s.nextBlock()

	if block == nil {
		t.Fatal("resume produced no audio")
	}
	if resumed := s.timeAt(at); resumed < heard-100*time.Millisecond || resumed > heard+100*time.Millisecond {
		t.Fatalf("resumed at %v but the last sound heard was %v: the queued audio was skipped",
			resumed, heard)
	}
}

// The needle does not move until the sound does: after a start the card sits silent while its
// cushion fills, and coasting on "not paused" races the needle ahead into that silence.
func TestTheTransportIsNotMovingUntilTheSoundIs(t *testing.T) {
	s := openFile(t)
	sink := &silentSink{}
	s.SetSink(sink)

	// unpaused, unparked - and still not making a sound, because the cushion is filling
	sink.on = false
	if s.Moving() {
		t.Fatal("the transport claims to be moving while the speaker is still filling its cushion")
	}

	sink.on = true
	if !s.Moving() {
		t.Fatal("the speaker is playing and the transport says it is not")
	}

	s.SetPaused(true)
	if s.Moving() {
		t.Fatal("a paused transport is not moving")
	}
}
