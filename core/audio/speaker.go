// core/audio/speaker.go
package audio

import (
	"math"
	"sync"
	"sync/atomic"

	"github.com/gen2brain/malgo"
)

// speakerPrefillSeconds is the cushion: how much audio must be queued before the card starts,
// and therefore how long the producer may be late before the sound breaks up. It is big on
// purpose - a small cushion demands the producer make a deadline every time while sharing the
// runtime with the DSP and the GC, and every miss is a hole. The needle is corrected against
// what the card actually played instead, so a deep queue costs nothing.
//
// speakerBufferSeconds is the ring itself, several times the cushion.
const (
	speakerBufferSeconds  = 1.5
	speakerPrefillSeconds = 0.35
)

// ring is the audio between the file and the card. Swapped, never resized, and only while the
// device is stopped, so the slice a callback reads is never the slice a Write is growing.
type ring struct {
	data []float32
}

// Speaker plays samples out of the default output device.
//
// THE CALLBACK TAKES NO LOCKS. pull runs on the sound card's realtime thread; if it waited on
// a mutex the producer held mid-copy, the driver would underrun. So the ring is a
// single-producer, single-consumer queue driven by two atomic cursors:
//
//	read   moved by the consumer (pull, on the audio thread)
//	write  moved by the producer (Write, on the source's goroutine)
//
// deviceMu still guards the DEVICE, which is not realtime work and the audio thread never
// touches. malgo's Uninit joins the audio thread, so a lock the callback wanted would deadlock.
type Speaker struct {
	deviceMu sync.Mutex
	ctx      *malgo.AllocatedContext
	device   *malgo.Device
	rate     int
	started  bool

	// Absolute sample counts, only ever going up. The modulo is taken at the point of use,
	// so "full" is write-read == size with no ambiguous empty-or-full state.
	buf   atomic.Pointer[ring]
	read  atomic.Int64
	write atomic.Int64

	muted atomic.Bool

	// Is the card actually consuming audio right now? NOT the same as "not paused": after a
	// start or resume it sits stopped until the cushion fills, and the needle must know or it
	// races ahead into silence and snaps back.
	playing atomic.Bool

	// gain holds math.Float64bits of 0..1. Atomic, read by the audio thread every buffer.
	gain atomic.Uint64

	// Faults, counted. Both are inaudible except as "a bit crunchy".
	underruns atomic.Int64 // the card asked for audio and the ring was empty
	drops     atomic.Int64 // audio arrived and the ring was full, so the oldest went
}

func NewSpeaker() (*Speaker, error) {
	ctx, err := newMalgoContext()
	if err != nil {
		return nil, err
	}
	sp := &Speaker{ctx: ctx}
	sp.SetVolume(1)
	return sp, nil
}

// Stats reports the two ways this can fail. Both stay at zero on a healthy machine.
func (s *Speaker) Stats() (underruns, drops int64) {
	return s.underruns.Load(), s.drops.Load()
}

// Played is how many samples the speaker has pushed at the card since the last reset. It only
// ever goes up - one counter, moved by one thread - which is why the needle is derived from it
// rather than from a difference that could move backwards.
func (s *Speaker) Played() int {
	n := s.read.Load()
	if n < 0 {
		return 0
	}
	return int(n)
}

// Buffered is how many samples were handed over but not played yet: what would be LOST if the
// queue were dropped. Independent of mute - a muted speaker plays exactly as much audio, so
// exactly as much waits to be heard. Only dropQueue needs it.
func (s *Speaker) Buffered() int {
	n := s.write.Load() - s.read.Load()
	if n < 0 {
		return 0
	}
	return int(n)
}

// Playing reports whether the card is actually pulling audio: device up and cushion filled.
func (s *Speaker) Playing() bool { return s.playing.Load() }

func (s *Speaker) Muted() bool { return s.muted.Load() }

// Volume is 0..1, applied to the SPEAKERS ONLY - it never touches a sample the engine sees,
// so it cannot quietly change the reading.
func (s *Speaker) Volume() float64 {
	return math.Float64frombits(s.gain.Load())
}

func (s *Speaker) SetVolume(v float64) {
	s.gain.Store(math.Float64bits(math.Max(0, math.Min(1, v))))
}

// SetMuted silences the room without touching the recording. It is a gain of zero and nothing
// else: the played-count is the clock the needle, the ghost mark and engine delivery all run
// on, so a muted speaker must consume its ring at the same rate a loud one does. Muting by
// refusing to feed the ring would stop that clock and desync the whole pipeline.
func (s *Speaker) SetMuted(muted bool) { s.muted.Store(muted) }
