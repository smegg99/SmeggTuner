package session

import (
	"smegg.me/smeggtuner/backend/common/i18n"
	"smegg.me/smeggtuner/backend/common/logger"
	coresession "smegg.me/smeggtuner/backend/core/session"
)

// SetA4 sets the reference pitch. Refused once there are readings: each is a deviation from the A4 it was taken at.
func (s *Service) SetA4(hz float64) error {
	if err := validA4(hz); err != nil {
		return err
	}
	s.mu.Lock()
	if s.active == nil {
		s.mu.Unlock()
		return ErrNoSession
	}
	if len(s.active.Takes) > 0 {
		s.mu.Unlock()
		logger.Warn(i18n.KeyLogSessionRejected, logger.Str("setting", "a4"), logger.Any("value", hz))
		return ErrHasReadings
	}
	s.active.A4 = hz
	s.mu.Unlock()

	s.touch()
	s.emitActive()
	return nil
}

// SetTranspose sets how many semitones above its key the instrument sounds, 0..11. Refused once
// there are readings, as A4 is: each was measured at the pitch the transposition made its key sound.
func (s *Service) SetTranspose(semitones int) error {
	if semitones < 0 || semitones > coresession.MaxTranspose {
		logger.Warn(i18n.KeyLogSessionRejected, logger.Str("setting", "transpose"), logger.Int("value", semitones))
		return ErrInvalidTranspose
	}
	s.mu.Lock()
	if s.active == nil {
		s.mu.Unlock()
		return ErrNoSession
	}
	if len(s.active.Takes) > 0 {
		s.mu.Unlock()
		logger.Warn(i18n.KeyLogSessionRejected, logger.Str("setting", "transpose"), logger.Int("value", semitones))
		return ErrHasReadings
	}
	s.active.Transpose = semitones
	s.mu.Unlock()

	s.touch()
	s.emitActive()
	return nil
}
