package session

import (
	"errors"

	"smegg.me/smeggtuner/common/logger"
	coresession "smegg.me/smeggtuner/core/session"
	"smegg.me/smeggtuner/core/target"
)

// FitCurve recovers the curve the instrument already holds from the session's readings and makes
// it the goal. The outliers travel back too: a bad reed must not bend the curve towards itself.
func (s *Service) FitCurve() (*FitDTO, error) {
	s.mu.Lock()
	defer s.unlockAndPublish()
	if s.active == nil {
		return nil, ErrNoSession
	}

	res, err := target.Fit(s.active.Readings(), s.active.Instrument.ReedCount, s.active.A4,
		target.FitOptions{Name: s.active.Name})
	switch {
	case errors.Is(err, target.ErrNoReadings):
		return nil, ErrNoReadings
	case err != nil:
		logger.Warn(logger.MsgSessionRejected, logger.Err(err))
		return nil, ErrFitFailed
	}

	s.active.Curve = res.Curve
	logger.Info(logger.MsgSessionCurveFitted,
		logger.Int("used", res.Used),
		logger.Int("merged", res.Merged), logger.Int("outliers", len(res.Outliers)))
	return &FitDTO{
		Curve:    cloneCurve(res.Curve),
		Outliers: res.Outliers,
		Used:     res.Used,
		Merged:   res.Merged,
	}, nil
}

// ImportCurve copies another session's goal onto this one. A curve wider or narrower than this
// instrument's reeds is taken as-is: a reed it says nothing about has no goal (zero).
func (s *Service) ImportCurve(fromID string) error {
	from, err := s.sessions().Get(fromID)
	if err != nil {
		logger.Warn(logger.MsgSessionLoadFailed, logger.Str("id", fromID), logger.Err(err))
		return loadError(err)
	}
	return s.importCurve(from, fromID)
}

// ImportCurveFile is ImportCurve from a session file anywhere on disk.
func (s *Service) ImportCurveFile(path string) error {
	from, err := coresession.ReadSessionAny(path)
	if err != nil {
		logger.Warn(logger.MsgSessionLoadFailed, logger.Str("path", path), logger.Err(err))
		return ErrLoadFailed
	}
	return s.importCurve(from, path)
}

func (s *Service) importCurve(from *coresession.Session, source string) error {
	if from.Curve == nil {
		return ErrNoCurve
	}
	s.mu.Lock()
	defer s.unlockAndPublish()
	if s.active == nil {
		return ErrNoSession
	}
	s.active.Curve = cloneCurve(from.Curve)
	logger.Info(logger.MsgSessionCurveImported,
		logger.Str("source", source), logger.Int("reeds", from.Curve.ReedCount),
		logger.Int("anchors", len(from.Curve.Anchors)))
	return nil
}
