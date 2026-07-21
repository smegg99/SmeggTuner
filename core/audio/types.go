// core/audio/types.go

// Package audio provides sample sources (mic, file, synthetic) and tone
// playback for the measurement core.
package audio

import (
	"context"
	"errors"
	"time"
)

// ErrDeviceLost reports that a capture device stopped on its own: unplugged, dropped, or
// taken away by the OS.
var ErrDeviceLost = errors.New("audio device lost")

// Block carries one chunk of mono samples. Samples are owned by the consumer (the engine
// filters them in place), so sources must not retain or reuse the slice.
type Block struct {
	Samples    []float32
	SampleRate int
	Time       time.Time

	// At is where in the recording these samples came from: the offset of the first of them.
	// Zero and meaningless for a mic. It rides with the audio so a measurement can say which
	// audio it was made from, since the engine's window trails the needle by over a second.
	At time.Duration
}

type SourceInfo struct {
	Name       string
	SampleRate int
	Realtime   bool
}

type Source interface {
	Start(ctx context.Context) (<-chan Block, error)
	Stop() error
	Info() SourceInfo
}

// Peak is one column of a drawn waveform: the lowest and highest sample in the stretch it
// stands for. Min and Max because the envelope is the information - the mean of a sinusoid's
// magnitude says nothing.
type Peak struct {
	Min float32
	Max float32
}

// Transport is a Source whose playhead can be driven: a file, and nothing else. The mic
// cannot be seeked, so the UI asks for this interface and simply draws no transport when the
// source lacks one.
type Transport interface {
	Duration() time.Duration
	Position() time.Duration
	Selection() (from, to time.Duration)
	Paused() bool

	// Moving is whether the playhead is ACTUALLY advancing - sound in the room this instant.
	// The needle coasts on this and nothing else; "not paused" makes it race through the
	// cushion-filling silence and snap back.
	Moving() bool

	Seek(at time.Duration)
	SetPaused(paused bool)
	SetRange(from, to time.Duration)
	SetLoop(loop bool)

	Peaks(from, to time.Duration, buckets int) []Peak
}

// Sink is somewhere audio goes on its way past the engine: the speakers. A FileSource pushes
// every block it hands the engine into its sink first, so what is heard is the SAME samples,
// at the same moment, the reeds on screen were measured from. An interface so the core owes
// nothing to a sound card: tests pass a recorder, a machine with no output passes nothing.
type Sink interface {
	// Write takes a block. It must not block the caller and must not keep the slice: the
	// engine filters these samples in place immediately afterwards.
	Write(samples []float32, sampleRate int)

	// Playing reports whether the sink is actually consuming audio right now. NOT "not
	// paused": after a start or resume it sits stopped until its cushion is full.
	Playing() bool

	// Played is how many samples have actually come out since the last Reset. It only ever
	// goes up, and the needle is derived from it. See Speaker.Played, FileSource.Position.
	Played() int

	// Buffered is how many samples were handed over but never heard - what would be lost if
	// the queue were dropped right now.
	Buffered() int

	// Reset discards what is queued and re-arms the cushion, so a fresh run starts full.
	// Called at every Start.
	Reset()

	Close() error
}

// ErrorSource is implemented by sources that can fail at runtime (the mic). After the block
// channel closes, Err reports why, or nil for a clean end. FileSource and SynthSource have a
// clean EOF and do not implement it.
type ErrorSource interface {
	Err() error
}

const blockSize = 4096

// fileBlockSize is the block a RECORDING is handed over in, small on purpose: a block reaches
// the engine only once the speaker has played it, so the newest audio a reading can describe
// trails the needle by up to one block. 1024 samples is ~21 ms, below what an eye finds on a
// waveform. The mic keeps the larger block - it has no timeline to line up.
const fileBlockSize = 1024
