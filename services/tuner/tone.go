package tuner

import (
	"math"
	"time"

	"smegg.me/smeggtuner/common/logger"
	coreaudio "smegg.me/smeggtuner/core/audio"
	"smegg.me/smeggtuner/core/tuning"
)

// PlayTone plays the note as a sine at the reference pitch, corrected for clock error.
// Sets toneUntil so observe() refuses to record while our own sine is in the room.
func (s *Service) PlayTone(note int) error {
	n := tuning.Note(note)
	if !n.Valid() {
		logger.Warn(logger.MsgTunerSettingRejected,
			logger.Str("setting", "tone_note"), logger.Int("value", note))
		return ErrInvalidNote
	}
	cfg := s.snapshot()
	player, err := s.tonePlayer()
	if err != nil {
		logger.Warn(logger.MsgTunerTonePlayFailed, logger.Err(err))
		return ErrPlaybackFailed
	}
	// A lazily opened player starts at full, so assert the level.
	player.SetVolume(s.toneVolume())
	// Set the deadline before Play so there is no window where the tone sounds but observe still trusts the room.
	dur := s.toneDur()
	s.toneUntil.Store(time.Now().Add(dur).UnixNano())

	if err := player.Play(n.Freq(cfg.A4), dur, cfg.ClockPPM); err != nil {
		s.toneUntil.Store(0)
		logger.Warn(logger.MsgTunerTonePlayFailed, logger.Err(err))
		return ErrPlaybackFailed
	}
	logger.Debug(logger.MsgTunerTonePlaying, logger.Int("note", note))
	return nil
}

// StopTone ends the reference tone; must be idempotent (also called on pointer-cancel and window blur).
func (s *Service) StopTone() error {
	s.toneUntil.Store(0)

	s.toneMu.Lock()
	player := s.tone
	s.toneMu.Unlock()

	if player != nil {
		player.Stop()
	}
	return nil
}

// SetToneVolume sets the tone level 0..1 and returns the clamped value; independent
// of file playback, applied on the callback's next buffer.
func (s *Service) SetToneVolume(v float64) float64 {
	if v < 0 {
		v = 0
	} else if v > 1 {
		v = 1
	}
	s.toneVolBits.Store(math.Float64bits(v))

	s.toneMu.Lock()
	player := s.tone
	s.toneMu.Unlock()
	if player != nil {
		player.SetVolume(v)
	}
	return v
}

// toneVolume is the current tone level 0..1; starts full, not persisted.
func (s *Service) toneVolume() float64 {
	return math.Float64frombits(s.toneVolBits.Load())
}

func (s *Service) tonePlaying() bool {
	until := s.toneUntil.Load()
	return until != 0 && time.Now().UnixNano() < until
}

// tonePlayer opens the playback device on first use and keeps it open.
func (s *Service) tonePlayer() (*coreaudio.TonePlayer, error) {
	s.toneMu.Lock()
	defer s.toneMu.Unlock()
	if s.tone != nil {
		return s.tone, nil
	}
	p, err := coreaudio.NewTonePlayer()
	if err != nil {
		return nil, err
	}
	s.tone = p
	return p, nil
}
