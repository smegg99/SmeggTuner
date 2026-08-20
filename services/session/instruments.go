package session

import (
	"errors"
	"os"
	"slices"

	"smegg.me/smeggtuner/common/logger"
	coresession "smegg.me/smeggtuner/core/session"
)

// Instruments lists every instrument a session can be started from.
func (s *Service) Instruments() ([]coresession.Template, error) {
	all, err := s.templates().List()
	if err != nil {
		logger.Warn(logger.MsgSessionLoadFailed, logger.Err(err))
		return nil, ErrLoadFailed
	}
	return all, nil
}

// DeleteInstrument removes a saved instrument; shipped ones cannot be deleted.
func (s *Service) DeleteInstrument(id string) error {
	if err := s.templates().Delete(id); err != nil {
		logger.Warn(logger.MsgSessionSaveFailed, logger.Str("id", id), logger.Err(err))
		return saveError(err)
	}
	return nil
}

// ImportInstrument reads an instrument file into the library under a new identity.
func (s *Service) ImportInstrument(path string) (*coresession.Template, error) {
	t, err := s.templates().ImportFile(path)
	if err != nil {
		logger.Warn(logger.MsgSessionLoadFailed, logger.Str("path", path), logger.Err(err))
		return nil, ErrLoadFailed
	}
	return t, nil
}

// ExportInstrument writes an instrument out to a file.
func (s *Service) ExportInstrument(id, path string) error {
	if err := s.templates().ExportFile(id, path); err != nil {
		logger.Warn(logger.MsgSessionSaveFailed, logger.Str("id", id), logger.Err(err))
		return saveError(err)
	}
	return nil
}

// SetInstrument rewrites the accordion on the bench. The card's columns - the ranks, the reed
// count, the bass ladder - freeze once there are readings, because every reading was filed under
// a column this would move and nobody re-measured them. Everything else stays open: a switch that
// had not been described yet, a tolerance, the keyboard's ends, the name.
func (s *Service) SetInstrument(i coresession.Instrument) error {
	if err := coresession.ValidInstrument(i); err != nil {
		logger.Warn(logger.MsgSessionRejected, logger.Str("setting", "instrument"), logger.Err(err))
		return ErrInvalidInstrument
	}

	s.mu.Lock()
	if s.active == nil {
		s.mu.Unlock()
		return ErrNoSession
	}
	if len(s.active.Takes) > 0 && columnsDiffer(s.active.Instrument, i) {
		s.mu.Unlock()
		logger.Warn(logger.MsgSessionRejected, logger.Str("setting", "instrument_columns"))
		return ErrHasReadings
	}
	s.active.Instrument = cloneInstrument(i)
	s.mu.Unlock()

	s.touch()
	s.emitActive()
	return nil
}

// columnsDiffer reports whether two descriptions would print a different card.
func columnsDiffer(a, b coresession.Instrument) bool {
	if a.ReedCount != b.ReedCount || a.BassReeds != b.BassReeds {
		return true
	}
	return !slices.Equal(a.Banks, b.Banks)
}

// SaveInstrumentSpec writes a remembered accordion; empty id creates one, an existing id edits it in place.
func (s *Service) SaveInstrumentSpec(t coresession.Template) (*coresession.Template, error) {
	if err := s.templates().Save(&t); err != nil {
		logger.Warn(logger.MsgSessionSaveFailed, logger.Err(err))
		return nil, saveError(err)
	}
	return &t, nil
}

// SetInstrumentImage puts a photograph on an instrument via core/session.PrepareImage; empty path removes it.
func (s *Service) SetInstrumentImage(id, path string) error {
	if path == "" {
		if err := s.templates().SetImage(id, nil); err != nil {
			logger.Warn(logger.MsgSessionSaveFailed, logger.Str("id", id), logger.Err(err))
			return saveError(err)
		}
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		logger.Warn(logger.MsgSessionLoadFailed, logger.Str("path", path), logger.Err(err))
		return ErrImageUnreadable
	}
	defer f.Close()

	jpg, err := coresession.PrepareImage(f)
	if err != nil {
		logger.Warn(logger.MsgSessionLoadFailed, logger.Str("path", path), logger.Err(err))
		return ErrNotAnImage
	}
	if err := s.templates().SetImage(id, jpg); err != nil {
		logger.Warn(logger.MsgSessionSaveFailed, logger.Str("id", id), logger.Err(err))
		return saveError(err)
	}
	return nil
}

// AdoptInstrument remembers a session's accordion as a template; it has no photograph.
// The open session is taken as it stands, not as it was last written: the store is written
// asynchronously (see touch and saver), so reading it back would remember whatever was
// there before the last keystroke.
func (s *Service) AdoptInstrument(sessionID, name string) (*coresession.Template, error) {
	sn := s.Snapshot()
	if sn == nil || sn.ID != sessionID {
		var err error
		if sn, err = s.sessions().Get(sessionID); err != nil {
			logger.Warn(logger.MsgSessionLoadFailed, logger.Str("id", sessionID), logger.Err(err))
			return nil, loadError(err)
		}
	}

	t := coresession.FromSession(sn, name)
	if err := s.templates().Save(t); err != nil {
		logger.Warn(logger.MsgSessionSaveFailed, logger.Err(err))
		return nil, saveError(err)
	}

	// Link the session to the new template so it never offers to adopt again. The open
	// session is linked in memory as well, or the next flush writes the link back off.
	sn.InstrumentID = t.ID
	if err := s.sessions().Save(sn); err != nil {
		logger.Warn(logger.MsgSessionSaveFailed, logger.Err(err))
		return nil, saveError(err)
	}

	s.mu.Lock()
	linked := s.active != nil && s.active.ID == sessionID
	if linked {
		s.active.InstrumentID = t.ID
	}
	s.mu.Unlock()

	if linked {
		s.emitActive()
	}
	return t, nil
}

// SuggestInstrumentFileName returns a suggested export filename for an instrument.
func (s *Service) SuggestInstrumentFileName(name string) string {
	return coresession.SuggestFileName(name)
}

func saveError(err error) error {
	switch {
	case errors.Is(err, coresession.ErrTemplateName):
		return ErrInstrumentName
	case errors.Is(err, coresession.ErrNoTemplate):
		return ErrNotFound
	}
	return ErrSaveFailed
}
