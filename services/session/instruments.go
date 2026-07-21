package session

import (
	"errors"
	"os"

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

// SaveInstrument keeps the active session's instrument as a named template, dropping the serial.
func (s *Service) SaveInstrument(name string) (*coresession.Template, error) {
	snap := s.Snapshot()
	if snap == nil {
		return nil, ErrNoSession
	}

	t := coresession.FromSession(snap, name)
	if err := s.templates().Save(t); err != nil {
		logger.Warn(logger.MsgSessionSaveFailed, logger.Err(err))
		return nil, saveError(err)
	}
	return t, nil
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

// SaveInstrumentSpec writes an instrument from the Instruments tab; empty id creates, an existing id edits in place.
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

// AdoptInstrument saves the session's recorded instrument as a template; it has no photograph.
func (s *Service) AdoptInstrument(sessionID, name string) (*coresession.Template, error) {
	sn, err := s.sessions().Get(sessionID)
	if err != nil {
		logger.Warn(logger.MsgSessionLoadFailed, logger.Str("id", sessionID), logger.Err(err))
		return nil, loadError(err)
	}

	t := coresession.FromSession(sn, name)
	if err := s.templates().Save(t); err != nil {
		logger.Warn(logger.MsgSessionSaveFailed, logger.Err(err))
		return nil, saveError(err)
	}

	// Link the session to the new template so it never offers to adopt again.
	sn.InstrumentID = t.ID
	if err := s.sessions().Save(sn); err != nil {
		logger.Warn(logger.MsgSessionSaveFailed, logger.Err(err))
		return nil, saveError(err)
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
