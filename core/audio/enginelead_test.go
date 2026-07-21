// core/audio/enginelead_test.go
package audio

import (
	"context"
	"sync"
	"testing"
	"time"
)

// The engine never runs ahead of the ears. A block used to reach the speaker and the engine at
// once, but the speaker plays it a cushion later - so the reading described audio nobody had
// heard. Now a block waits until the speaker has played it.
func TestTheEngineNeverRunsAheadOfTheEars(t *testing.T) {
	s, err := NewFileSource(fixture, true, false)
	if err != nil {
		t.Fatal(err)
	}
	sink := &cushionSink{rate: 48000}
	s.SetSink(sink)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	blocks, err := s.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// a card, eating audio at real time once it has started
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		tick := time.NewTicker(10 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-stop:
				return
			case <-tick.C:
				if sink.Playing() {
					sink.play(480)
				}
			}
		}
	}()

	var got int
	deadline := time.After(2 * time.Second)

	for got < 12 {
		select {
		case b, ok := <-blocks:
			if !ok {
				t.Fatal("the stream ended early")
			}
			got++

			// where this block ENDS in the recording: the newest audio the engine now has
			end := b.At + time.Duration(float64(len(b.Samples))/48000*float64(time.Second))

			if heard := s.Position(); end > heard+50*time.Millisecond {
				t.Fatalf("the engine was handed audio at %v while the room had only heard %v: "+
					"the reading describes sound nobody has heard, and pausing leaves it "+
					"standing in the future", end, heard)
			}
		case <-deadline:
			t.Fatalf("the engine was starved: only %d blocks in two seconds", got)
		}
	}
}

// The engine is given the audio that is in the room, not what finished playing a moment ago.
// The producer used to deliver only after sleeping its pacing interval, and to wait for a
// block to be FULLY played - so the block coming out of the speakers had never been delivered.
func TestTheEngineIsGivenTheAudioThatIsInTheRoom(t *testing.T) {
	s, err := NewFileSource(fixture, true, true)
	if err != nil {
		t.Fatal(err)
	}
	sink := &cushionSink{rate: 48000}
	s.SetSink(sink)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	blocks, err := s.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// a card, eating audio at real time once it has started
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		tick := time.NewTicker(5 * time.Millisecond)
		defer tick.Stop()
		for {
			select {
			case <-stop:
				return
			case <-tick.C:
				if sink.Playing() {
					sink.play(240) // 5 ms at 48 kHz
				}
			}
		}
	}()

	var delivered time.Duration
	var mu sync.Mutex
	go func() {
		for b := range blocks {
			end := b.At + time.Duration(float64(len(b.Samples))/48000*float64(time.Second))
			mu.Lock()
			delivered = end
			mu.Unlock()
		}
	}()

	time.Sleep(time.Second)

	for i := 0; i < 3; i++ {
		s.SetPaused(true)
		time.Sleep(120 * time.Millisecond) // everything settles

		needle := s.Position()
		mu.Lock()
		got := delivered
		mu.Unlock()

		// The engine may hold slightly MORE than the room has heard (it is given the block the
		// speaker has started) but never be short of it by more than a block.
		if short := needle - got; short > 40*time.Millisecond {
			t.Fatalf("the speaker had played up to %v and the engine had only been given %v: "+
				"%v of audio in the room that no reading can be about", needle, got, short)
		}

		s.SetPaused(false)
		time.Sleep(200 * time.Millisecond)
	}
}
