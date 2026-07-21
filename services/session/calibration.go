package session

import (
	"smegg.me/smeggtuner/common/logger"
	coresession "smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/tuning"
)

// SetKeyboardRange sets the keyboard's low/high MIDI notes (zero = not learned). It also writes
// the range to the shelf instrument, and is allowed while a pass is open (no reading depends on it).
func (s *Service) SetKeyboardRange(lo, hi int) error {
	s.mu.Lock()
	if s.active == nil {
		s.mu.Unlock()
		return ErrNoSession
	}

	next := s.active.Instrument
	next.Lo = tuning.Note(lo)
	next.Hi = tuning.Note(hi)
	if err := coresession.ValidInstrument(next); err != nil {
		s.mu.Unlock()
		logger.Warn(logger.MsgSessionRejected, logger.Str("setting", "range"), logger.Err(err))
		return ErrInvalidInstrument
	}

	s.active.Instrument = next
	id := s.active.InstrumentID
	s.mu.Unlock()

	s.rememberRange(id, tuning.Note(lo), tuning.Note(hi))

	s.touch()
	s.emitActive()
	return nil
}

// rememberRange writes the range to the shelf instrument; no shelf copy is not a failure.
func (s *Service) rememberRange(id string, lo, hi tuning.Note) {
	if id == "" {
		return
	}
	t, err := s.templates().Get(id)
	if err != nil {
		return
	}
	t.Instrument.Lo, t.Instrument.Hi = lo, hi
	if err := s.templates().Save(t); err != nil {
		logger.Warn(logger.MsgSessionSaveFailed, logger.Str("id", id), logger.Err(err))
	}
}
