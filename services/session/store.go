package session

import (
	"smegg.me/smeggtuner/common/logger"
	"smegg.me/smeggtuner/core/repositories"
	coresession "smegg.me/smeggtuner/core/session"
)

// adopt makes next the active session, ending whatever was open before it.
func (s *Service) adopt(next *coresession.Session) (*SessionDTO, error) {
	if err := s.Close(); err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.active = next
	// A new instrument gets its own bench; nothing carries over.
	s.register = s.defaultRegisterLocked()
	dto := s.dtoLocked()
	s.mu.Unlock()

	s.emitActive()
	return dto, nil
}

// unlockAndPublish releases the write lock, schedules a save and tells the frontend.
func (s *Service) unlockAndPublish() {
	s.mu.Unlock()
	s.touch()
	s.emitActive()
}

// dtoLocked renders the active session for the frontend. Callers hold mu.
func (s *Service) dtoLocked() *SessionDTO {
	if s.active == nil {
		return nil
	}
	a := s.active
	dto := &SessionDTO{
		ID:           a.ID,
		Name:         a.Name,
		Instrument:   cloneInstrument(a.Instrument),
		InstrumentID: a.InstrumentID,
		A4:           a.A4,
		Curve:        cloneCurve(a.Curve),
		Readings:     len(a.Takes),
		Notes:        a.Notes,
		Bench:        s.benchLocked(),
		Created:      a.Created,
		Updated:      a.Updated,
	}
	return dto
}

func (s *Service) emitActive() {
	s.mu.RLock()
	dto := s.dtoLocked()
	s.mu.RUnlock()
	emitEvent(EventActive, ActiveDTO{Session: dto})
}

// sessions and templates are stateless repositories fetched per call; nothing to build or guard.
func (s *Service) sessions() *repositories.SessionRepository {
	return repositories.GetSessionRepository()
}

func (s *Service) templates() *repositories.InstrumentRepository {
	return repositories.GetInstrumentRepository()
}

// touch schedules a save; never blocks. A request that finds one queued is dropped, since the saver writes the latest state.
func (s *Service) touch() {
	select {
	case s.dirty <- struct{}{}:
	default:
	}
}

// saver writes the session out when touched, off the emit path so an fsync never stalls measurements.
func (s *Service) saver() {
	defer close(s.done)
	for {
		select {
		case <-s.dirty:
			if err := s.flush(); err != nil {
				emitEvent(EventSaveFailed, ErrorDTO{Key: err.Error()})
			}
		case <-s.quit:
			return
		}
	}
}

// flush writes the active session to disk. Snapshot under the read lock, written outside it; saveMu
// orders flushes so an older cannot land on top of a newer.
func (s *Service) flush() error {
	s.saveMu.Lock()
	defer s.saveMu.Unlock()

	s.mu.RLock()
	if s.active == nil {
		s.mu.RUnlock()
		return nil
	}
	snap := cloneSession(s.active)
	s.mu.RUnlock()

	if err := s.sessions().Save(snap); err != nil {
		logger.Warn(logger.MsgSessionSaveFailed, logger.Str("id", snap.ID), logger.Err(err))
		return ErrSaveFailed
	}

	// Save stamps the write time on the copy; copy it back or the list shows it older than its file.
	s.mu.Lock()
	if s.active != nil && s.active.ID == snap.ID {
		s.active.Updated = snap.Updated
	}
	s.mu.Unlock()
	return nil
}
