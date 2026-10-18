package session

import (
	"math"

	"smegg.me/smeggtuner/backend/common/i18n"

	"smegg.me/smeggtuner/backend/common/logger"
	coresession "smegg.me/smeggtuner/backend/core/session"
	"smegg.me/smeggtuner/backend/core/target"
	"smegg.me/smeggtuner/backend/core/tuning"
)

// Goal is what the active session imposes on the engine; the zero Goal when none is open.
func (s *Service) Goal() coresession.Goal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.active == nil {
		return coresession.Goal{}
	}
	bench := s.benchLocked()
	return coresession.Goal{
		Curve:         s.active.Curve,
		A4:            s.active.A4,
		Transpose:     s.active.Transpose,
		Reeds:         bench.Reeds,
		Banks:         bench.Banks,
		CardBanks:     s.active.Instrument.Banks,
		BassFeet:      bench.BassFeet,
		CardFeet:      coresession.BassFeet(s.active.Instrument.BassReeds),
		Tolerance:     s.active.Instrument.Tolerance,
		BeatTolerance: s.active.Instrument.BeatTolerance,
	}
}

// SetAnchor pins one reed of one note of the goal curve; value is read in unit ("reference", "cent",
// or "hz" at the note's pitch for A4 = 440 Hz, whatever the session's A4).
func (s *Service) SetAnchor(note, reed int, value float64, unit string) error {
	u, err := parseUnit(unit)
	if err != nil {
		return err
	}
	if !tuning.Note(note).Valid() {
		return ErrInvalidNote
	}

	s.mu.Lock()
	defer s.unlockAndPublish()
	if s.active == nil {
		return ErrNoSession
	}
	c := s.editCurveLocked()
	if reed < 0 || reed >= c.ReedCount {
		return ErrInvalidReed
	}
	c.Unit = u // the entry's unit decides how value converts to stored reference units
	if err := c.Set(tuning.Note(note), reed, value); err != nil {
		logger.Warn(i18n.KeyLogSessionRejected, logger.Err(err))
		return ErrInvalidValue
	}
	s.active.Curve = c
	return nil
}

// SetBeating writes a whole note of the goal curve from one beating value, the beat between reeds 1
// and 2, in unit as SetAnchor reads it; RefReed sits at pitch and Asymmetry spreads the pairs above.
func (s *Service) SetBeating(note int, value float64, unit string) error {
	u, err := parseUnit(unit)
	if err != nil {
		return err
	}
	if !tuning.Note(note).Valid() {
		return ErrInvalidNote
	}

	s.mu.Lock()
	defer s.unlockAndPublish()
	if s.active == nil {
		return ErrNoSession
	}
	c := s.editCurveLocked()
	c.Unit = u
	if err := c.SetBeating(tuning.Note(note), value); err != nil {
		logger.Warn(i18n.KeyLogSessionRejected, logger.Err(err))
		return ErrInvalidValue
	}
	s.active.Curve = c
	return nil
}

// SetAsymmetry makes each pair the next SetBeating writes beat 1 + percent/100 times the pair below
// it, -100..100. It does not re-derive anchors already entered.
func (s *Service) SetAsymmetry(percent float64) error {
	if math.IsNaN(percent) || percent < -target.MaxAsymmetry || percent > target.MaxAsymmetry {
		logger.Warn(i18n.KeyLogSessionRejected,
			logger.Str("setting", "asymmetry"), logger.Any("value", percent))
		return ErrInvalidValue
	}

	s.mu.Lock()
	defer s.unlockAndPublish()
	if s.active == nil {
		return ErrNoSession
	}
	c := s.editCurveLocked()
	c.Asymmetry = percent
	s.active.Curve = c
	return nil
}

// ClearAnchor drops the anchor at a note, leaving an empty curve if it was the last.
func (s *Service) ClearAnchor(note int) error {
	s.mu.Lock()
	defer s.unlockAndPublish()
	if s.active == nil {
		return ErrNoSession
	}
	if s.active.Curve == nil {
		return nil
	}
	c := cloneCurve(s.active.Curve)
	c.Clear(tuning.Note(note))
	s.active.Curve = c
	return nil
}

// DropCurve removes the goal, leaving the tuner a pure indicator.
func (s *Service) DropCurve() error {
	s.mu.Lock()
	defer s.unlockAndPublish()
	if s.active == nil {
		return ErrNoSession
	}
	s.active.Curve = nil
	logger.Info(i18n.KeyLogSessionCurveDropped)
	return nil
}

// editCurveLocked returns a clone of the active curve, or a new instrument-wide one; always a copy
// (see clone.go). Callers hold mu.
func (s *Service) editCurveLocked() *target.Curve {
	if c := cloneCurve(s.active.Curve); c != nil {
		return c
	}
	return target.NewCurve(s.active.Name, s.active.Instrument.ReedCount)
}
