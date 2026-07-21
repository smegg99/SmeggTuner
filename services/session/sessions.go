package session

import (
	"errors"

	"smegg.me/smeggtuner/common/logger"
	"smegg.me/smeggtuner/core/repositories"
	coresession "smegg.me/smeggtuner/core/session"
)

// List summarizes every session, most-recently-touched first; unparseable files are skipped, not fatal.
func (s *Service) List() ([]coresession.Summary, error) {
	out, err := s.sessions().List()
	if err != nil {
		logger.Warn(logger.MsgSessionLoadFailed, logger.Err(err))
		return nil, ErrLoadFailed
	}
	if out == nil {
		out = []coresession.Summary{}
	}
	return out, nil
}

// Active returns the open session, or nil when none is.
func (s *Service) Active() *SessionDTO {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dtoLocked()
}

// Snapshot returns the active session deep-copied under one lock, or nil when none is open.
// services/report needs the copy: a take appended mid-read cannot alter the model it is drawing.
func (s *Service) Snapshot() *coresession.Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.active == nil {
		return nil
	}
	return cloneSession(s.active)
}

// Create makes a session for an instrument and opens it.
func (s *Service) Create(dto NewSessionDTO) (*SessionDTO, error) {
	if err := validName(dto.Name); err != nil {
		return nil, err
	}
	// A4 rides on the instrument; an undescribed one inherits the default.
	a4 := dto.Instrument.A4
	if a4 == 0 {
		a4 = defaultA4
	}
	if err := validA4(a4); err != nil {
		return nil, err
	}
	if err := validReeds(dto.Instrument.ReedCount); err != nil {
		return nil, err
	}
	next := coresession.New(dto.Name, cloneInstrument(dto.Instrument), a4)
	next.Notes = dto.Notes
	// A link, not a dependency: the instrument itself was copied in above.
	next.InstrumentID = dto.InstrumentID
	if err := next.Validate(); err != nil {
		logger.Warn(logger.MsgSessionRejected, logger.Err(err))
		return nil, ErrInvalidReedCount
	}
	if err := s.sessions().Save(next); err != nil {
		logger.Warn(logger.MsgSessionSaveFailed, logger.Err(err))
		return nil, ErrSaveFailed
	}
	logger.Info(logger.MsgSessionCreated,
		logger.Str("id", next.ID), logger.Str("name", next.Name),
		logger.Any("a4", next.A4), logger.Int("reeds", next.Instrument.ReedCount))
	return s.adopt(next)
}

// Open loads a session and makes it active, ending whatever was open before. An instrument that
// sounds more reeds than the engine resolves opens unchanged; the clamp is services/tuner's.
func (s *Service) Open(id string) (*SessionDTO, error) {
	next, err := s.sessions().Get(id)
	if err != nil {
		logger.Warn(logger.MsgSessionLoadFailed, logger.Str("id", id), logger.Err(err))
		return nil, loadError(err)
	}
	logger.Info(logger.MsgSessionOpened,
		logger.Str("id", next.ID), logger.Str("name", next.Name),
		logger.Any("a4", next.A4), logger.Int("reeds", next.Instrument.ReedCount),
		logger.Int("readings", len(next.Takes)))
	return s.adopt(next)
}

// Close writes the active session out and lets go of it.
func (s *Service) Close() error {
	s.mu.RLock()
	none := s.active == nil
	s.mu.RUnlock()
	if none {
		return nil
	}

	err := s.flush()

	s.mu.Lock()
	if s.active != nil {
		logger.Info(logger.MsgSessionClosed, logger.Str("id", s.active.ID))
	}
	s.active = nil
	s.mu.Unlock()

	s.emitActive()
	return err
}

// Save writes the active session out now and reports the result.
func (s *Service) Save() error { return s.flush() }

// Delete removes a session from the store, closing it first if it is the one open.
func (s *Service) Delete(id string) error {
	s.mu.RLock()
	isActive := s.active != nil && s.active.ID == id
	s.mu.RUnlock()
	if isActive {
		if err := s.Close(); err != nil {
			return err
		}
	}
	if err := s.sessions().Delete(id); err != nil {
		logger.Warn(logger.MsgSessionSaveFailed, logger.Str("id", id), logger.Err(err))
		if errors.Is(err, repositories.ErrNotFound) {
			return ErrNotFound
		}
		return ErrSaveFailed
	}
	logger.Info(logger.MsgSessionDeleted, logger.Str("id", id))
	return nil
}

// SetA4 sets the reference pitch. Refused once there are readings: each is cents from the A4 it was taken at.
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
		logger.Warn(logger.MsgSessionRejected, logger.Str("setting", "a4"), logger.Any("value", hz))
		return ErrHasReadings
	}
	s.active.A4 = hz
	s.mu.Unlock()

	s.touch()
	s.emitActive()
	return nil
}
