// core/audio/transport_test.go
package audio

import (
	"context"
	"testing"
	"time"
)

const fixture = "../../tests/fixtures/h-16.wav"

func openFile(t *testing.T) *FileSource {
	t.Helper()
	s, err := NewFileSource(fixture, false, false)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestAFreshFileIsSelectedWhole(t *testing.T) {
	s := openFile(t)

	if s.Duration() <= 0 {
		t.Fatalf("duration %v", s.Duration())
	}
	from, to := s.Selection()
	if from != 0 || to != s.Duration() {
		t.Fatalf("selection %v..%v, want the whole file 0..%v", from, to, s.Duration())
	}
	if s.Position() != 0 {
		t.Fatalf("playhead starts at %v, want 0", s.Position())
	}
}

// The needle may not leave the selection - the guarantee the file view is built on, made in
// the producer rather than by clamping a drag in the canvas.
func TestTheNeedleCannotLeaveTheSelection(t *testing.T) {
	s := openFile(t)
	s.SetRange(2*time.Second, 4*time.Second)

	for _, seek := range []time.Duration{
		0,
		time.Second,              // before the selection
		3 * time.Second,          // inside it
		9 * time.Second,          // past it
		s.Duration(),             // the very end of the file
		s.Duration() + time.Hour, // and past the end of the file
		-time.Hour,               // and before the start of time
	} {
		s.Seek(seek)

		pos := s.Position()
		if pos < 2*time.Second || pos >= 4*time.Second {
			t.Fatalf("seek to %v put the playhead at %v, outside the selection 2s..4s", seek, pos)
		}
	}
}

// A selection made behind the playhead drags the playhead back with it.
func TestSelectingBehindThePlayheadBringsItBack(t *testing.T) {
	s := openFile(t)
	s.Seek(8 * time.Second)
	s.SetRange(1*time.Second, 2*time.Second)

	if pos := s.Position(); pos < time.Second || pos >= 2*time.Second {
		t.Fatalf("playhead left at %v after selecting 1s..2s behind it", pos)
	}
}

// A click with no drag is not a request to mute the transport.
func TestAnEmptySelectionIsTheWholeFile(t *testing.T) {
	s := openFile(t)

	for _, r := range [][2]time.Duration{
		{3 * time.Second, 3 * time.Second}, // empty
		{4 * time.Second, 2 * time.Second}, // dragged backwards
	} {
		s.SetRange(r[0], r[1])

		from, to := s.Selection()
		if from != 0 || to != s.Duration() {
			t.Fatalf("SetRange(%v, %v) gave %v..%v, want the whole file", r[0], r[1], from, to)
		}
	}
}

// The engine must never see a sample from outside the selection.
func TestPlaybackNeverLeavesTheSelection(t *testing.T) {
	s := openFile(t)
	s.SetLoop(true)
	s.SetRange(2*time.Second, 3*time.Second)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	blocks, err := s.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 200; i++ {
		select {
		case b, ok := <-blocks:
			if !ok {
				t.Fatal("the stream ended; a looping source has no end")
			}
			_ = b
		case <-time.After(2 * time.Second):
			t.Fatal("the source stopped producing audio inside its selection")
		}

		if pos := s.Position(); pos < 2*time.Second || pos > 3*time.Second {
			t.Fatalf("the playhead reached %v, outside the selection 2s..3s", pos)
		}
	}
}

// FileSource is the only Source that can be driven, and the UI asks for the capability.
func TestAFileIsATransport(t *testing.T) {
	var _ Transport = openFile(t)
}

func TestPeaksDrawTheEnvelope(t *testing.T) {
	s := openFile(t)

	const buckets = 400
	peaks := s.Peaks(0, s.Duration(), buckets)
	if len(peaks) != buckets {
		t.Fatalf("got %d peaks, want %d", len(peaks), buckets)
	}

	// Every column is a real min/max pair, and the file is not silent.
	var loudest float32
	for i, p := range peaks {
		if p.Min > p.Max {
			t.Fatalf("peak %d has min %v above max %v", i, p.Min, p.Max)
		}
		if p.Max > loudest {
			loudest = p.Max
		}
	}
	if loudest <= 0 {
		t.Fatal("the whole waveform is flat; h-16.wav is not silent")
	}

	// Zoomed in past one sample per column, the envelope still fills every column.
	if tight := s.Peaks(time.Second, time.Second+time.Millisecond, 500); len(tight) != 500 {
		t.Fatalf("zoomed in, got %d peaks, want 500", len(tight))
	}

	// and an inside-out or empty request draws nothing rather than panicking
	if p := s.Peaks(3*time.Second, time.Second, 100); p != nil {
		t.Fatalf("a backwards range gave %d peaks, want none", len(p))
	}
	if p := s.Peaks(0, s.Duration(), 0); p != nil {
		t.Fatalf("zero buckets gave %d peaks, want none", len(p))
	}
}
