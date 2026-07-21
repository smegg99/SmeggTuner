// core/audio/sink_test.go
package audio

import (
	"context"
	"testing"
	"time"
)

// What you hear is what was measured: the sink is fed the same slice the engine gets, and fed
// it FIRST, because the engine filters in place. It is also the only claim that makes
// listening useful - the fragment on loop is the audio the numbers came out of.
func TestYouHearTheSamplesTheEngineMeasures(t *testing.T) {
	s, err := NewFileSource(fixture, true, false) // realtime is the only path the speakers are on
	if err != nil {
		t.Fatal(err)
	}
	s.SetRange(1*time.Second, 2*time.Second)

	sink := &recordingSink{}
	s.SetSink(sink)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	blocks, err := s.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Drained until the source PARKS, not until the channel closes: a non-looping selection
	// plays out and sits with the stream open (it is about to be seeked back into).
	var measured []float32
drain:
	for {
		select {
		case b, ok := <-blocks:
			if !ok {
				break drain
			}
			measured = append(measured, b.Samples...)
		case <-time.After(500 * time.Millisecond):
			break drain // parked: the selection has played out
		}
	}

	heard, rate := sink.all()
	if len(heard) == 0 {
		t.Fatal("nothing reached the speakers")
	}
	if rate != s.sampleRate {
		t.Fatalf("played at %d Hz, but the recording is %d Hz: nothing may resample it", rate, s.sampleRate)
	}
	if len(heard) != len(measured) {
		t.Fatalf("heard %d samples but measured %d", len(heard), len(measured))
	}
	for i := range heard {
		if heard[i] != measured[i] {
			t.Fatalf("sample %d: heard %v, measured %v", i, heard[i], measured[i])
		}
	}

	// and only the selection was audible
	if want := s.sampleRate; len(heard) < want-fileBlockSize || len(heard) > want+fileBlockSize {
		t.Fatalf("heard %d samples of a one second selection, want about %d", len(heard), want)
	}
}

// A restart leaves one producer, not two: a second writer corrupts the lock-free speaker ring.
// Run under -race.
func TestARestartLeavesOneProducer(t *testing.T) {
	s, err := NewFileSource(fixture, true, true) // realtime, looping
	if err != nil {
		t.Fatal(err)
	}

	sink := &recordingSink{}
	s.SetSink(sink)

	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		blocks, err := s.Start(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// let it get going, then restart it out from under itself
		<-blocks
		cancel()
	}

	// and one last run that is allowed to live, so a straggler has every chance to still be
	// writing alongside it
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	blocks, err := s.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		<-blocks
	}
}

// Every run gets a full cushion, not just the first: a Stop drains the speaker but left the
// device started, so the next play skipped the prefill and broke up from the first block.
func TestEveryStartRearmsTheSpeaker(t *testing.T) {
	s, err := NewFileSource(fixture, true, false)
	if err != nil {
		t.Fatal(err)
	}
	sink := &resettingSink{}
	s.SetSink(sink)

	for i := 1; i <= 3; i++ {
		before := sink.count()

		ctx, cancel := context.WithCancel(context.Background())
		blocks, err := s.Start(ctx)
		if err != nil {
			t.Fatal(err)
		}

		// The count is taken at Start, not after the run (stopping also flushes): what is
		// asserted is that STARTING re-arms the cushion.
		if sink.count() <= before {
			t.Fatalf("start %d did not re-arm the speaker, so it begins with no cushion", i)
		}

		<-blocks
		cancel()
	}
}

// Pause is immediate: the speaker holds a cushion of handed-over-but-unheard audio, and merely
// stopping the producer would let it play out, so pause would land late by the cushion's size.
func TestPauseThrowsAwayWhatIsQueued(t *testing.T) {
	s, err := NewFileSource(fixture, true, false)
	if err != nil {
		t.Fatal(err)
	}
	sink := &resettingSink{}
	s.SetSink(sink)

	before := sink.count()
	s.SetPaused(true)
	if sink.count() != before+1 {
		t.Fatal("pause left the queued audio to play out; it must be dropped")
	}

	// and resuming does not throw anything away - there is nothing to throw
	at := sink.count()
	s.SetPaused(false)
	if sink.count() != at {
		t.Fatal("resume flushed the speaker, which would drop the cushion it just built")
	}
}
