package session

import (
	"errors"
	"unicode/utf8"

	"smegg.me/smeggtuner/backend/common/i18n"

	"smegg.me/smeggtuner/backend/common/logger"
	"smegg.me/smeggtuner/backend/core/repositories"
	coresession "smegg.me/smeggtuner/backend/core/session"
)

// List summarizes stored sessions, most-recently-touched first.
func (s *Service) List() ([]coresession.Summary, error) {
	out, err := s.sessions().List()
	if err != nil {
		logger.Warn(i18n.KeyLogSessionLoadFailed, logger.Err(err))
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
	s.opMu.Lock()
	defer s.opMu.Unlock()
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
		logger.Warn(i18n.KeyLogSessionRejected, logger.Err(err))
		return nil, ErrInvalidReedCount
	}
	if err := s.sessions().Save(next); err != nil {
		logger.Warn(i18n.KeyLogSessionSaveFailed, logger.Err(err))
		return nil, ErrSaveFailed
	}
	logger.Info(i18n.KeyLogSessionCreated,
		logger.Str("id", next.ID), logger.Str("name", next.Name),
		logger.Any("a4", next.A4), logger.Int("reeds", next.Instrument.ReedCount))
	return s.adopt(next)
}

// Open loads a session and makes it active, ending whatever was open before. An instrument that
// sounds more reeds than the engine resolves opens unchanged; the clamp is services/tuner's.
func (s *Service) Open(id string) (*SessionDTO, error) {
	s.opMu.Lock()
	defer s.opMu.Unlock()
	if current := s.Active(); current != nil && current.ID == id {
		return current, nil
	}
	next, err := s.sessions().Get(id)
	if err != nil {
		logger.Warn(i18n.KeyLogSessionLoadFailed, logger.Str("id", id), logger.Err(err))
		return nil, loadError(err)
	}
	logger.Info(i18n.KeyLogSessionOpened,
		logger.Str("id", next.ID), logger.Str("name", next.Name),
		logger.Any("a4", next.A4), logger.Int("reeds", next.Instrument.ReedCount),
		logger.Int("readings", len(next.Takes)))
	return s.adopt(next)
}

// Close writes the active session out and lets go of it.
func (s *Service) Close() error {
	s.opMu.Lock()
	defer s.opMu.Unlock()
	return s.close()
}

func (s *Service) close() error {
	s.saveMu.Lock()
	s.mu.Lock()
	if s.active == nil {
		s.mu.Unlock()
		s.saveMu.Unlock()
		return nil
	}
	// Keep the exact state being closed locked until it has been persisted.
	snap := cloneSession(s.active)
	if err := s.sessions().Save(snap); err != nil {
		s.mu.Unlock()
		s.saveMu.Unlock()
		logger.Warn(i18n.KeyLogSessionSaveFailed, logger.Str("id", snap.ID), logger.Err(err))
		return ErrSaveFailed
	}
	logger.Info(i18n.KeyLogSessionClosed, logger.Str("id", snap.ID))
	s.active = nil
	s.mu.Unlock()
	s.saveMu.Unlock()

	s.emitActive()
	return nil
}

// Save writes the active session out now and reports the result.
func (s *Service) Save() error { return s.flush() }

// Delete removes a session from the store, closing it first if it is the one open.
func (s *Service) Delete(id string) error {
	s.opMu.Lock()
	defer s.opMu.Unlock()
	s.mu.RLock()
	isActive := s.active != nil && s.active.ID == id
	s.mu.RUnlock()
	if isActive {
		if err := s.close(); err != nil {
			return err
		}
	}
	if err := s.sessions().Delete(id); err != nil {
		logger.Warn(i18n.KeyLogSessionSaveFailed, logger.Str("id", id), logger.Err(err))
		if errors.Is(err, repositories.ErrNotFound) {
			return ErrNotFound
		}
		return ErrSaveFailed
	}
	logger.Info(i18n.KeyLogSessionDeleted, logger.Str("id", id))
	return nil
}

// maxNotes bounds Unicode code points in the note printed under the card's table. The dialog caps it too,
// but this is where the guarantee has to hold.
const maxNotes = 500

// SetNotes rewrites the session's note. Unlike the reference pitch this stays open
// with readings present: the note describes the work, and the work is what produced them.
func (s *Service) SetNotes(text string) error {
	if utf8.RuneCountInString(text) > maxNotes {
		logger.Warn(i18n.KeyLogSessionRejected, logger.Str("setting", "notes"), logger.Int("value", len(text)))
		return ErrNotesTooLong
	}
	s.mu.Lock()
	if s.active == nil {
		s.mu.Unlock()
		return ErrNoSession
	}
	s.active.Notes = text
	s.mu.Unlock()

	s.touch()
	s.emitActive()
	return nil
}
