package session

import (
	"math"

	"smegg.me/smeggtuner/backend/common/i18n"

	"smegg.me/smeggtuner/backend/common/logger"
	coresession "smegg.me/smeggtuner/backend/core/session"
)

// Data returns the session's readings with the reference, instrument, and curve needed to read them.
func (s *Service) Data() (*ReadingData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.active == nil {
		return nil, ErrNoSession
	}
	return &ReadingData{
		SessionID:  s.active.ID,
		A4:         s.active.A4,
		Transpose:  s.active.Transpose,
		ReedCount:  s.active.Instrument.ReedCount,
		Instrument: cloneInstrument(s.active.Instrument),
		Curve:      s.active.Curve,
		Takes:      cloneTakes(s.active.Takes),
	}, nil
}

// UpsertTake records a reading, replacing that voice's previous one. On the emit path, so it does not wait for disk.
func (s *Service) UpsertTake(t coresession.Take) error {
	s.mu.Lock()
	if s.active == nil {
		s.mu.Unlock()
		return ErrNoSession
	}
	s.stampLocked(&t)
	s.active.UpsertTake(t)
	s.mu.Unlock()

	s.touch()
	s.emitActive()
	return nil
}

// UndoTake drops the reading captured last and reports whether there was one; an empty session is not an error.
func (s *Service) UndoTake() (bool, error) {
	s.mu.Lock()
	if s.active == nil {
		s.mu.Unlock()
		return false, ErrNoSession
	}
	_, ok := s.active.UndoLast()
	s.mu.Unlock()

	if ok {
		s.touch()
		s.emitActive()
	}
	return ok, nil
}

// DeleteTake removes one reading by its index.
func (s *Service) DeleteTake(i int) error {
	s.mu.Lock()
	if s.active == nil {
		s.mu.Unlock()
		return ErrNoSession
	}
	if !s.active.DeleteTake(i) {
		s.mu.Unlock()
		return ErrTakeNotFound
	}
	s.mu.Unlock()

	logger.Info(i18n.KeyLogRecordUndone, logger.Int("reading", i))
	s.touch()
	s.emitActive()
	return nil
}

// ClearTakes removes all readings, keeping the session and its curve.
func (s *Service) ClearTakes() error {
	s.mu.Lock()
	if s.active == nil {
		s.mu.Unlock()
		return ErrNoSession
	}
	s.active.Clear()
	s.mu.Unlock()

	s.touch()
	s.emitActive()
	return nil
}

// SetTakeReed writes a sounding slot by hand in reference units at the session's A4 and transposition, and marks the take manual.
func (s *Service) SetTakeReed(take, reed int, dev float64) error {
	if math.IsNaN(dev) || math.IsInf(dev, 0) {
		return ErrInvalidValue
	}
	s.mu.Lock()
	defer s.unlockAndPublish()
	if s.active == nil {
		return ErrNoSession
	}
	if take < 0 || take >= len(s.active.Takes) {
		return ErrTakeNotFound
	}
	if !s.active.Takes[take].SetReed(reed, dev, s.active.A4, s.active.Transpose) {
		return ErrInvalidReed
	}
	return nil
}
