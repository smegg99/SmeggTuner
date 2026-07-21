// core/audio/playback_test.go
package audio

import (
	"context"
	"testing"
	"time"
)

// Pause stops the audio without ending the stream: a paused file is about to be resumed.
func TestPauseStopsTheAudioAndKeepsTheStream(t *testing.T) {
	s := openFile(t)
	s.SetPaused(true)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	blocks, err := s.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case b, ok := <-blocks:
		if !ok {
			t.Fatal("pause closed the stream; it must stay open to be resumed")
		}
		t.Fatalf("a paused source emitted %d samples", len(b.Samples))
	case <-time.After(150 * time.Millisecond):
	}

	at := s.Position()
	s.SetPaused(false)

	select {
	case _, ok := <-blocks:
		if !ok {
			t.Fatal("the stream ended instead of resuming")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("resume produced no audio")
	}
	if s.Position() <= at {
		t.Fatalf("the playhead did not advance on resume: %v then %v", at, s.Position())
	}
}

// The headless path (whole file, no loop, no pause) must still end, or every batch analysis
// would hang on a channel that never closes.
func TestAWholeFileStillEnds(t *testing.T) {
	s := openFile(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	blocks, err := s.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}

	var n int
	done := make(chan struct{})
	go func() {
		defer close(done)
		for b := range blocks {
			n += len(b.Samples)
		}
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("the stream never closed")
	}
	if want := int(s.Duration().Seconds() * 48000); n < want-fileBlockSize || n > want+fileBlockSize {
		t.Fatalf("played %d samples, want about %d", n, want)
	}
}

// Pressing play on a recording that has run out plays it again. The source is kept between
// runs, so its playhead is parked at the end; a Start that did not rewind would emit nothing.
func TestStartingAFinishedFilePlaysItAgain(t *testing.T) {
	s := openFile(t)
	s.SetRange(1*time.Second, 2*time.Second)
	s.Seek(2 * time.Second) // played out: the playhead is parked at the end

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	blocks, err := s.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}

	select {
	case b, ok := <-blocks:
		if !ok || len(b.Samples) == 0 {
			t.Fatal("a finished file produced no audio on the next Start")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("a finished file produced no audio on the next Start")
	}

	// and it rewound to the selection, not to the top of the file
	if pos := s.Position(); pos < time.Second || pos > 2*time.Second {
		t.Fatalf("it resumed at %v, outside the selection 1s..2s", pos)
	}
}

// A fragment that has played out goes quiet and stays quiet. It used to park on a sample still
// inside the selection, so it dribbled a single sample every block forever - it hung a test.
func TestAPlayedOutSelectionGoesQuiet(t *testing.T) {
	s, err := NewFileSource(fixture, true, false)
	if err != nil {
		t.Fatal(err)
	}
	s.SetRange(1*time.Second, time.Duration(1.3*float64(time.Second)))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	blocks, err := s.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// let the fragment play itself out
	deadline := time.After(3 * time.Second)
played:
	for {
		select {
		case <-blocks:
		case <-time.After(400 * time.Millisecond):
			break played
		case <-deadline:
			break played
		}
	}

	// and now nothing more may come out of it
	select {
	case b, ok := <-blocks:
		if ok {
			t.Fatalf("a played-out fragment emitted another %d samples; it must stay parked", len(b.Samples))
		}
	case <-time.After(500 * time.Millisecond):
	}

	if pos := s.Position(); pos < time.Second || pos > time.Duration(1.3*float64(time.Second)) {
		t.Fatalf("the needle parked at %v, outside the fragment it just played", pos)
	}
}

// A paused file keeps its place. Only a FINISHED one rewinds.
func TestStartingAPausedFileKeepsItsPlace(t *testing.T) {
	s := openFile(t)
	s.Seek(4 * time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if _, err := s.Start(ctx); err != nil {
		t.Fatal(err)
	}

	if pos := s.Position(); pos < 4*time.Second {
		t.Fatalf("Start rewound a file that had not finished: %v, want 4s or past it", pos)
	}
}

// A batch analysis rips through a file as fast as it decodes; pointing that at a card would
// play at forty times speed. The headless path pushes nothing at the speakers.
func TestAHeadlessRunIsSilent(t *testing.T) {
	s, err := NewFileSource(fixture, false, false) // realtime = false
	if err != nil {
		t.Fatal(err)
	}

	sink := &recordingSink{}
	s.SetSink(sink)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	blocks, err := s.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for range blocks { //nolint:revive // draining is the point
	}

	if heard, _ := sink.all(); len(heard) != 0 {
		t.Fatalf("a headless run pushed %d samples at the speakers", len(heard))
	}
}
