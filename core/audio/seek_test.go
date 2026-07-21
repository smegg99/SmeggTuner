// core/audio/seek_test.go
package audio

import (
	"testing"
	"time"
)

// A seek moves the needle now, not a cushion from now. Position is a lookup over the marks, so
// while the queue still holds the audio from before the seek it kept answering with the old
// spot. Clearing the marks makes Position read the playhead, where the seek just put it.
func TestASeekMovesTheNeedleAtOnce(t *testing.T) {
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

	if heard := s.Position(); heard < 3*time.Second || heard > 4*time.Second {
		t.Fatalf("the needle is at %v, which is not where this test put it", heard)
	}

	s.Seek(time.Second)

	if got := s.Position(); got < 900*time.Millisecond || got > 1100*time.Millisecond {
		t.Fatalf("seeked to 1s but the needle reads %v: the queued audio still owns the "+
			"playhead, so the seek is invisible until it drains", got)
	}
}

// And the audio follows the needle: the next block comes from where it was seeked to.
func TestASeekPlaysFromWhereItLanded(t *testing.T) {
	s := openFile(t)
	sink := &drainingSink{}
	s.SetSink(sink)

	for handed := 0; handed < 4*48000; {
		block, at, _ := s.nextBlock()
		if block == nil {
			break
		}
		s.markBlock(at, len(block))
		sink.Write(block, 48000)
		handed += len(block)
	}
	sink.play(175_200)

	s.Seek(time.Second)

	block, at, _ := s.nextBlock()
	if block == nil {
		t.Fatal("the seek produced no audio")
	}
	if resumed := s.timeAt(at); resumed < 900*time.Millisecond || resumed > 1100*time.Millisecond {
		t.Fatalf("seeked to 1s but the next audio is from %v", resumed)
	}
}

// A selection made behind the needle brings the needle with it, visibly - SetRange drags the
// playhead but used to move it the same silent way Seek did.
func TestSelectingBehindTheNeedleMovesItAtOnce(t *testing.T) {
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

	// the fragment is behind the needle, so the needle has to come back to it
	s.SetRange(0, time.Second)

	if got := s.Position(); got > time.Second {
		t.Fatalf("selected 0..1s but the needle reads %v, outside the fragment: the queued "+
			"audio still owns the playhead", got)
	}
}

// Selecting AROUND the needle leaves it exactly where it was: the sound has not changed, so
// neither may the needle pointing at it.
func TestSelectingAroundTheNeedleLeavesItAlone(t *testing.T) {
	s := openFile(t)
	sink := &drainingSink{}
	s.SetSink(sink)

	for handed := 0; handed < 4*48000; {
		block, at, _ := s.nextBlock()
		if block == nil {
			break
		}
		s.markBlock(at, len(block))
		sink.Write(block, 48000)
		handed += len(block)
	}
	sink.play(175_200)

	before := s.Position()
	s.SetRange(3*time.Second, 5*time.Second) // the needle at 3.65 is already inside

	if got := s.Position(); got < before-50*time.Millisecond || got > before+50*time.Millisecond {
		t.Fatalf("the needle moved from %v to %v on a selection that contains it", before, got)
	}
}
