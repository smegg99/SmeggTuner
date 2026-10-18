package session

import (
	"errors"

	"smegg.me/smeggtuner/backend/common/i18n"

	"smegg.me/smeggtuner/backend/common/logger"
	coresession "smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/target"
)

// FitCurve recovers the curve the instrument already holds from the session's readings and makes
// it the goal. The outliers travel back too: a bad reed must not bend the curve towards itself.
func (s *Service) FitCurve() (*FitDTO, error) {
	s.mu.Lock()
	defer s.unlockAndPublish()
	if s.active == nil {
		return nil, ErrNoSession
	}

	width := s.active.Instrument.ReedCount
	res, err := target.Fit(s.active.Readings(width), width, target.FitOptions{Name: s.active.Name})
	switch {
	case errors.Is(err, target.ErrNoReadings):
		return nil, ErrNoReadings
	case err != nil:
		logger.Warn(i18n.KeyLogSessionRejected, logger.Err(err))
		return nil, ErrFitFailed
	}

	s.active.Curve = res.Curve
	logger.Info(i18n.KeyLogSessionCurveFitted,
		logger.Int("used", res.Used), logger.Int("outliers", len(res.Outliers)))
	return &FitDTO{
		Curve:    cloneCurve(res.Curve),
		Outliers: res.Outliers,
		Used:     res.Used,
	}, nil
}

// ImportCurve copies another session's goal onto this one, replacing it: the reference's "load beat
// list". A curve wider or narrower than this instrument's reeds is taken as-is: a column it says
// nothing about has no goal. A session without anchors has nothing to give (ErrNoCurve).
func (s *Service) ImportCurve(fromID string) error {
	from, err := s.sessions().Get(fromID)
	if err != nil {
		logger.Warn(i18n.KeyLogSessionLoadFailed, logger.Str("id", fromID), logger.Err(err))
		return loadError(err)
	}
	return s.importCurve(from, fromID)
}

// ImportCurveFile is ImportCurve from a session file anywhere on disk.
func (s *Service) ImportCurveFile(path string) error {
	from, err := coresession.ReadSessionFile(path)
	if err != nil {
		logger.Warn(i18n.KeyLogSessionLoadFailed, logger.Str("path", path), logger.Err(err))
		return ErrLoadFailed
	}
	return s.importCurve(from, path)
}

func (s *Service) importCurve(from *coresession.Session, source string) error {
	if from.Curve == nil || len(from.Curve.Anchors) == 0 {
		return ErrNoCurve
	}
	s.mu.Lock()
	defer s.unlockAndPublish()
	if s.active == nil {
		return ErrNoSession
	}
	s.active.Curve = cloneCurve(from.Curve)
	logger.Info(i18n.KeyLogSessionCurveImported,
		logger.Str("source", source), logger.Int("reeds", from.Curve.ReedCount),
		logger.Int("anchors", len(from.Curve.Anchors)))
	return nil
}
