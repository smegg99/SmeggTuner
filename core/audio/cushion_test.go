// core/audio/cushion_test.go
package audio

import (
	"context"
	"testing"
	"time"
)

// Play is immediate: the cushion is filled at once, not at the speed it will be heard. The gap
// after Play was never the SIZE of the cushion, it was the time spent filling it at real time.
func TestTheCushionIsFilledAtOnceSoPlayIsImmediate(t *testing.T) {
	s, err := NewFileSource(fixture, true, false) // realtime
	if err != nil {
		t.Fatal(err)
	}
	sink := &cushionSink{rate: 48000}
	s.SetSink(sink)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	started := time.Now()
	blocks, err := s.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		for range blocks { //nolint:revive
		}
	}()

	for !sink.Playing() {
		if time.Since(started) > 200*time.Millisecond {
			t.Fatalf("the speaker was still silent %v after Play: the cushion is being filled "+
				"at the speed it will be heard, which is the gap", time.Since(started))
		}
		time.Sleep(time.Millisecond)
	}

	// and it does not then sprint through the recording: the lead is spent, not compounded
	if lead := s.leadOverRealtime(started); lead > 900*time.Millisecond {
		t.Fatalf("it ran %v ahead of real time: filling the cushion must not become a sprint", lead)
	}
}

// The cushion survives being built and survives a resume. Every block handed over early also
// pushed the pacing deadline forward, so once the card started the producer slept it all off
// and the card ran dry. With the lead no longer carried forward it holds near its target.
func TestTheCushionSurvivesTheBurstAndAResume(t *testing.T) {
	s, err := NewFileSource(fixture, true, true) // realtime, looping
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
	go func() {
		for range blocks { //nolint:revive
		}
	}()

	// A card: it eats audio at exactly real time, and only once it has started.
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
					sink.play(480) // 10 ms at 48 kHz
				}
			}
		}
	}()

	// watch returns the smallest cushion seen while the card was actually pulling
	watch := func(d time.Duration) int {
		lowest := 1 << 30
		deadline := time.Now().Add(d)
		for time.Now().Before(deadline) {
			if sink.Playing() {
				if b := sink.Buffered(); b < lowest {
					lowest = b
				}
			}
			time.Sleep(2 * time.Millisecond)
		}
		return lowest
	}

	// Generous: the point is that it does not collapse to nothing.
	const floor = 48000 / 20 // 50 ms

	if low := watch(time.Second); low < floor {
		t.Fatalf("the cushion fell to %.0f ms on the first play: it is being spent as fast "+
			"as it is built, and the card runs dry", float64(low)/48)
	}

	s.SetPaused(true)
	time.Sleep(200 * time.Millisecond)
	s.SetPaused(false)

	if low := watch(time.Second); low < floor {
		t.Fatalf("the cushion fell to %.0f ms after a resume: playback breaks up every time "+
			"it is restarted", float64(low)/48)
	}
}
