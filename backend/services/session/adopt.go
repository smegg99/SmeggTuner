package session

import (
	"smegg.me/smeggtuner/backend/common/i18n"
	"smegg.me/smeggtuner/backend/common/logger"
	coresession "smegg.me/smeggtuner/backend/core/session"
)

// AdoptInstrument remembers a session's accordion as a template; it has no photograph.
// The open session is taken as it stands, not as it was last written: the store is written
// asynchronously (see touch and saver), so reading it back would remember whatever was
// there before the last keystroke.
func (s *Service) AdoptInstrument(sessionID, name string) (*coresession.Template, error) {
	s.opMu.Lock()
	defer s.opMu.Unlock()
	s.saveMu.Lock()
	s.mu.Lock()
	t, linked, err := s.adoptInstrumentLocked(sessionID, name)
	s.mu.Unlock()
	s.saveMu.Unlock()
	if err == nil && linked {
		s.emitActive()
	}
	return t, err
}

func (s *Service) adoptInstrumentLocked(sessionID, name string) (*coresession.Template, bool, error) {
	linked := s.active != nil && s.active.ID == sessionID
	var sn *coresession.Session
	if linked {
		sn = cloneSession(s.active)
	} else {
		var err error
		if sn, err = s.sessions().Get(sessionID); err != nil {
			logger.Warn(i18n.KeyLogSessionLoadFailed, logger.Str("id", sessionID), logger.Err(err))
			return nil, false, loadError(err)
		}
	}

	t := coresession.FromSession(sn, name)
	if err := s.templates().Save(t); err != nil {
		logger.Warn(i18n.KeyLogSessionSaveFailed, logger.Err(err))
		return nil, false, saveError(err)
	}
	if err := s.sessions().LinkInstrument(sessionID, t.ID); err != nil {
		logger.Warn(i18n.KeyLogSessionSaveFailed, logger.Err(err))
		return nil, false, saveError(err)
	}
	if linked {
		s.active.InstrumentID = t.ID
	}
	return t, linked, nil
}
