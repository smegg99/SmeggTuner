// core/audio/speaker_test.go
package audio

import (
	"encoding/binary"
	"math"
	"testing"
)

// A muted speaker still keeps time: the played-count is the clock the needle, ghost mark and
// engine delivery run on, so mute must be a gain of zero and nothing else. These say so.

// loaded is a speaker with a ring full of audio and no sound card, which is all pull needs.
func loaded(t *testing.T, samples int) *Speaker {
	t.Helper()

	sp := &Speaker{}
	sp.SetVolume(1)

	data := make([]float32, samples)
	for i := range data {
		data[i] = 0.5 // audible, so silence is a fact and not a coincidence
	}
	sp.buf.Store(&ring{data: data})
	sp.write.Store(int64(samples))
	return sp
}

// pullFrames runs one callback and hands back what the card would have heard.
func pullFrames(sp *Speaker, frames int) []float32 {
	out := make([]byte, frames*4)
	sp.pull(out, nil, uint32(frames))

	got := make([]float32, frames)
	for i := range got {
		got[i] = math.Float32frombits(binary.LittleEndian.Uint32(out[i*4:]))
	}
	return got
}

// The regression: a muted speaker consumes its ring at the same rate a loud one does.
func TestAMutedSpeakerStillKeepsTime(t *testing.T) {
	const frames = 256

	loud := loaded(t, 4096)
	pullFrames(loud, frames)

	quiet := loaded(t, 4096)
	quiet.SetMuted(true)
	pullFrames(quiet, frames)

	if quiet.Played() != loud.Played() {
		t.Fatalf("a muted speaker played %d samples and a loud one %d: the clock stops when the sound does",
			quiet.Played(), loud.Played())
	}
	if quiet.Played() != frames {
		t.Fatalf("played = %d, want %d", quiet.Played(), frames)
	}
}

// And it is genuinely silent while it does it.
func TestAMutedSpeakerIsSilent(t *testing.T) {
	sp := loaded(t, 1024)
	sp.SetMuted(true)

	for _, v := range pullFrames(sp, 256) {
		if v != 0 {
			t.Fatalf("a muted speaker put %v out of the card", v)
		}
	}

	// Unmuted, the same ring is audible - so the silence above was the mute, not an empty buffer.
	sp.SetMuted(false)
	heard := false
	for _, v := range pullFrames(sp, 256) {
		if v != 0 {
			heard = true
		}
	}
	if !heard {
		t.Fatal("the ring was empty, so the test above proved nothing")
	}
}

// Unmuting does not release a hoard of stale audio: what comes out next is what is under the
// needle now, because the ring was consumed while muted.
func TestUnmutingPlaysTheAudioUnderTheNeedleAndNotThePast(t *testing.T) {
	sp := loaded(t, 4096)
	sp.SetMuted(true)
	pullFrames(sp, 1024)

	was := sp.Played()
	sp.SetMuted(false)
	pullFrames(sp, 256)

	if sp.Played() != was+256 {
		t.Fatalf("played %d, want %d: unmuting rewound the queue", sp.Played(), was+256)
	}
	if sp.Buffered() != 4096-was-256 {
		t.Fatalf("buffered = %d: a mute's worth of audio was held back", sp.Buffered())
	}
}

// Buffered is what would be lost if the queue were dropped, and a mute does not change that -
// dropQueue owes the playhead the same samples whether loud or quiet.
func TestAMutedSpeakerOwesThePlayheadTheSameAudio(t *testing.T) {
	sp := loaded(t, 4096)
	sp.SetMuted(true)
	pullFrames(sp, 1024)

	if got, want := sp.Buffered(), 4096-1024; got != want {
		t.Fatalf("buffered = %d, want %d", got, want)
	}
}

// The volume knob moves no cursor: it is applied per sample on the way out.
func TestTheVolumeKnobDoesNotTouchTheClock(t *testing.T) {
	sp := loaded(t, 4096)

	pullFrames(sp, 256)
	sp.SetVolume(0.25)
	pullFrames(sp, 256)
	sp.SetVolume(1)
	pullFrames(sp, 256)

	if sp.Played() != 768 {
		t.Fatalf("played = %d, want 768: turning the knob moved the playhead", sp.Played())
	}
}
