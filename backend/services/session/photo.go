package session

import (
	"os"

	"smegg.me/smeggtuner/backend/common/i18n"
	"smegg.me/smeggtuner/backend/common/logger"
	coresession "smegg.me/smeggtuner/backend/core/session"
)

// SetInstrumentImage puts a photograph on an instrument via core/session.PrepareImage; empty path removes it.
func (s *Service) SetInstrumentImage(id, path string) error {
	if path == "" {
		if err := s.templates().SetImage(id, nil); err != nil {
			logger.Warn(i18n.KeyLogSessionSaveFailed, logger.Str("id", id), logger.Err(err))
			return saveError(err)
		}
		return nil
	}

	f, err := os.Open(path)
	if err != nil {
		logger.Warn(i18n.KeyLogSessionLoadFailed, logger.Str("path", path), logger.Err(err))
		return ErrImageUnreadable
	}
	defer func() { _ = f.Close() }()

	jpg, err := coresession.PrepareImage(f)
	if err != nil {
		logger.Warn(i18n.KeyLogSessionLoadFailed, logger.Str("path", path), logger.Err(err))
		return ErrNotAnImage
	}
	if err := s.templates().SetImage(id, jpg); err != nil {
		logger.Warn(i18n.KeyLogSessionSaveFailed, logger.Str("id", id), logger.Err(err))
		return saveError(err)
	}
	return nil
}
