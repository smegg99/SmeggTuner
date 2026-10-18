package session

import (
	"errors"
	"slices"

	"smegg.me/smeggtuner/backend/common/i18n"

	"smegg.me/smeggtuner/backend/common/logger"
	coresession "smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/internal/documents"
)

// Instruments lists every instrument a session can be started from.
func (s *Service) Instruments() ([]coresession.Template, error) {
	all, err := s.templates().List()
	if err != nil {
		logger.Warn(i18n.KeyLogSessionLoadFailed, logger.Err(err))
		return nil, ErrLoadFailed
	}
	return all, nil
}

// DeleteInstrument removes a saved instrument.
func (s *Service) DeleteInstrument(id string) error {
	if err := s.templates().Delete(id); err != nil {
		logger.Warn(i18n.KeyLogSessionSaveFailed, logger.Str("id", id), logger.Err(err))
		return saveError(err)
	}
	return nil
}

// ImportInstrument reads an instrument file into the library under a new identity.
func (s *Service) ImportInstrument(path string) (*coresession.Template, error) {
	t, err := s.templates().ImportFile(path)
	if err != nil {
		logger.Warn(i18n.KeyLogSessionLoadFailed, logger.Str("path", path), logger.Err(err))
		return nil, ErrLoadFailed
	}
	return t, nil
}

// ExportInstrument writes an instrument out to a file.
func (s *Service) ExportInstrument(id, path string) error {
	if err := documents.Write(path, func(local string) error {
		return s.templates().ExportFile(id, local)
	}); err != nil {
		logger.Warn(i18n.KeyLogSessionSaveFailed, logger.Str("id", id), logger.Err(err))
		return saveError(err)
	}
	return nil
}

// SetInstrument updates the named active session's accordion and its reference before recording.
func (s *Service) SetInstrument(sessionID string, i coresession.Instrument) error {
	if err := coresession.ValidInstrument(i); err != nil {
		logger.Warn(i18n.KeyLogSessionRejected, logger.Str("setting", "instrument"), logger.Err(err))
		return ErrInvalidInstrument
	}
	if i.A4 != 0 {
		if err := validA4(i.A4); err != nil {
			return err
		}
	}

	s.mu.Lock()
	if s.active == nil {
		s.mu.Unlock()
		return ErrNoSession
	}
	if s.active.ID != sessionID {
		s.mu.Unlock()
		return ErrNoSession
	}
	if len(s.active.Takes) > 0 && (columnsDiffer(s.active.Instrument, i) || takesDiffer(s.active, i) || (i.A4 != 0 && i.A4 != s.active.A4)) {
		s.mu.Unlock()
		logger.Warn(i18n.KeyLogSessionRejected, logger.Str("setting", "instrument_columns"))
		return ErrHasReadings
	}
	s.active.Instrument = cloneInstrument(i)
	if i.A4 != 0 {
		s.active.A4 = i.A4
	}
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

// takesDiffer rejects edits that change the ranks behind an existing reading.
func takesDiffer(sn *coresession.Session, next coresession.Instrument) bool {
	for _, take := range sn.Takes {
		if take.Bass {
			if !slices.Equal(sn.Instrument.SoundingFeet(take.Register), next.SoundingFeet(take.Register)) {
				return true
			}
		} else if !slices.Equal(sn.Instrument.TakeBanks(take), next.TakeBanks(take)) {
			return true
		}
	}
	copy := *sn
	copy.Instrument = next
	return copy.Validate() != nil
}

// SaveInstrumentSpec writes a remembered accordion; empty id creates one, an existing id edits it in place.
func (s *Service) SaveInstrumentSpec(t coresession.Template) (*coresession.Template, error) {
	if err := s.templates().Save(&t); err != nil {
		logger.Warn(i18n.KeyLogSessionSaveFailed, logger.Err(err))
		return nil, saveError(err)
	}
	return &t, nil
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
