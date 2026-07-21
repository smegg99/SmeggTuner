// core/audio/filetransport.go
package audio

import (
	"time"
)

// Duration is the length of the whole file, not of the selection.
func (s *FileSource) Duration() time.Duration {
	return s.timeAt(len(s.samples))
}

// Position is which sample of the recording is coming out of the speaker right now. It is a
// LOOKUP against the speaker's played count, not a subtraction of two counters sampled at
// different instants (which glitched backwards by up to a block and broke on a loop wrap).
func (s *FileSource) Position() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Nobody listening (headless): the playhead is simply where we have read to.
	if s.sink == nil {
		return s.timeAt(s.pos)
	}

	played := s.sink.Played()

	// Everything handed over has been played, or nothing was: the two agree.
	if len(s.marks) == 0 {
		return s.timeAt(s.pos)
	}

	// Drop the blocks the speaker has finished with; played only goes up.
	for len(s.marks) > 1 && played >= s.marks[0].played+s.marks[0].n {
		s.marks = s.marks[1:]
	}

	m := s.marks[0]
	if played < m.played {
		return s.timeAt(m.file) // the card has not reached this block yet
	}

	into := played - m.played
	if into > m.n {
		into = m.n // it has run past the last block we gave it
	}
	return s.timeAt(m.file + into)
}

// Selection is the range playback is confined to.
func (s *FileSource) Selection() (from, to time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.timeAt(s.from), s.timeAt(s.to)
}

func (s *FileSource) Paused() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.paused
}

// Moving is whether sound is coming out of the speakers this instant - not "is it unpaused".
// The needle coasts on this; coasting on "not paused" makes it race ahead through the silence
// while the speaker fills its cushion, then snap back.
func (s *FileSource) Moving() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.paused || s.parked {
		return false
	}
	if s.sink == nil {
		return true // nobody is listening, so nothing waits on a speaker
	}
	return s.sink.Playing()
}

// SetPaused stops the audio now. The speaker holds a cushion of handed-over-but-unheard
// audio; merely stopping the producer would let it play out, so pause would land late by the
// size of the cushion. Dropping the queue makes it immediate, and resume re-prefills it.
func (s *FileSource) SetPaused(paused bool) {
	s.mu.Lock()
	s.paused = paused
	s.mu.Unlock()

	if paused {
		s.dropQueue()
	}
}

// dropQueue throws away the audio the speaker has not played yet and gives the playhead back
// the samples that were never heard. Without the second half the needle would jump forward by
// the cushion, and resuming would silently skip that unheard audio.
func (s *FileSource) dropQueue() {
	s.mu.Lock()
	sink := s.sink
	s.mu.Unlock()

	if sink == nil {
		return
	}

	unheard := sink.Buffered()
	sink.Reset()

	s.mu.Lock()
	s.pos = clampInt(s.pos-unheard, s.from, s.lastPlayable())
	s.forgetMarks()
	s.gen++ // the blocks the producer is holding back are about to be produced again
	s.mu.Unlock()
}

func (s *FileSource) SetLoop(loop bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loop = loop
	// Switching looping on at the end of a fragment is a request to hear it again.
	if loop {
		s.parked = false
	}
}

// Seek moves the playhead. It cannot leave the selection (see the type's doc); a seek past
// the end lands on the last playable sample.
func (s *FileSource) Seek(at time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pos = clampInt(s.sampleAt(at), s.from, s.lastPlayable())
	s.parked = false

	// The marks describe audio from before the seek, so they must not answer for the needle
	// after it: with no marks Position reads the playhead directly, which is where the seek
	// just put it. The queue is deliberately NOT dropped (that would stop the card, and a
	// drag seeks on every pointer move); the queued audio plays out and later blocks re-mark
	// normally.
	s.marks = nil
}

// SetRange confines playback to [from, to). The playhead comes with it, or a selection made
// behind the needle would leave the needle outside the only audio the engine may see. A
// backwards or empty selection is the whole file (what a click with no drag produces).
func (s *FileSource) SetRange(from, to time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	f := clampInt(s.sampleAt(from), 0, len(s.samples))
	t := clampInt(s.sampleAt(to), 0, len(s.samples))
	if t <= f {
		f, t = 0, len(s.samples)
	}
	s.from, s.to = f, t

	// Only clear the marks if the fragment actually DRAGGED the playhead (see Seek). A
	// selection that merely contains the needle changes nothing about the sound, and
	// clearing the marks there would shove the needle forward onto s.pos, a cushion ahead.
	if moved := clampInt(s.pos, s.from, s.lastPlayable()); moved != s.pos {
		s.pos = moved
		s.marks = nil
	}

	s.parked = false // a new fragment has not been played yet
}
