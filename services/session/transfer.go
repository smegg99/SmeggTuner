package session

import (
	"smegg.me/smeggtuner/common/logger"
	coresession "smegg.me/smeggtuner/core/session"
)

// ImportSession reads a session from a file under a new identity, so it cannot overwrite a local one. Saved, not opened.
func (s *Service) ImportSession(path string) (*coresession.Summary, error) {
	in, err := coresession.ReadSessionAny(path)
	if err != nil {
		logger.Warn(logger.MsgSessionLoadFailed, logger.Str("path", path), logger.Err(err))
		return nil, ErrLoadFailed
	}

	in.ID = coresession.NewID()
	if err := s.sessions().Save(in); err != nil {
		logger.Warn(logger.MsgSessionSaveFailed, logger.Err(err))
		return nil, saveError(err)
	}

	sum := in.Summarize()
	return &sum, nil
}

// ExportSession writes a session out whole, flushing the active one first so the export matches disk.
func (s *Service) ExportSession(id, path string) error {
	if active := s.Snapshot(); active != nil && active.ID == id {
		if err := s.flush(); err != nil {
			return err
		}
	}

	out, err := s.sessions().Get(id)
	if err != nil {
		logger.Warn(logger.MsgSessionLoadFailed, logger.Str("id", id), logger.Err(err))
		return loadError(err)
	}
	if err := coresession.WriteSessionFile(path, out); err != nil {
		logger.Warn(logger.MsgSessionSaveFailed, logger.Str("id", id), logger.Err(err))
		return saveError(err)
	}
	return nil
}

// SuggestSessionFileName returns a suggested export filename for a session.
func (s *Service) SuggestSessionFileName(name string) string {
	return coresession.SuggestSessionFileName(name)
}
