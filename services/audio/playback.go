package audio

import (
	"smegg.me/smeggtuner/common/logger"
	coreaudio "smegg.me/smeggtuner/core/audio"
)

// sinkOrNil keeps a typed nil out of the Sink interface: a (*Speaker)(nil) in a Sink is not a nil Sink and panics on Write.
func sinkOrNil(sp *coreaudio.Speaker) coreaudio.Sink {
	if sp == nil {
		return nil
	}
	return sp
}

// speakerLocked returns the output sink, opened once, or nil if none. Caller holds mu.
func (s *Service) speakerLocked() *coreaudio.Speaker {
	if s.speakerTried {
		return s.speaker
	}
	s.speakerTried = true

	sp, err := coreaudio.NewSpeaker()
	if err != nil {
		// Warn only: playback is optional, measuring still works without a speaker.
		logger.Warn(logger.MsgAudioSpeakerUnavailable, logger.Err(err))
		return nil
	}
	s.speaker = sp
	return sp
}

// SetMuted silences file playback without touching the transport.
func (s *Service) SetMuted(muted bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sp := s.speakerLocked(); sp != nil {
		sp.SetMuted(muted)
		return sp.Muted()
	}
	return true
}

// SetVolume scales the speakers only, 0..1; the engine gets the recording at its own level.
func (s *Service) SetVolume(v float64) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sp := s.speakerLocked(); sp != nil {
		sp.SetVolume(v)
		return sp.Volume()
	}
	return v
}

func (s *Service) Volume() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.speaker == nil {
		return 1
	}
	return s.speaker.Volume()
}

func (s *Service) Muted() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.speaker == nil {
		return s.speakerTried // untried: assume it will work and start unmuted
	}
	return s.speaker.Muted()
}

// Close releases the output device.
func (s *Service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.speaker == nil {
		return nil
	}
	err := s.speaker.Close()
	s.speaker = nil
	return err
}
