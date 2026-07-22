package session

import (
	"math"

	"smegg.me/smeggtuner/common/logger"
	"smegg.me/smeggtuner/core/target"
	"smegg.me/smeggtuner/core/tuning"
)

// Goal is what the active session imposes on the engine; the zero Goal when none is open.
func (s *Service) Goal() Goal {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.active == nil {
		return Goal{}
	}
	bench := s.benchLocked()
	return Goal{
		Curve:         s.active.Curve,
		A4:            s.active.A4,
		Reeds:         bench.Reeds,
		Banks:         bench.Banks,
		Profile:       s.active.Profile(),
		ProfileRev:    s.active.ProfileRev(),
		Tolerance:     s.active.Instrument.Tolerance,
		BeatTolerance: s.active.Instrument.BeatTolerance,
	}
}

// SetAnchor pins one reed of one note of the goal curve; value is read in unit ("cent" or "hz").
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
	c.Unit = u // the entry's unit decides how value converts to stored cents
	if err := c.Set(tuning.Note(note), reed, value, s.active.A4); err != nil {
		logger.Warn(logger.MsgSessionRejected, logger.Err(err))
		return ErrInvalidValue
	}
	s.active.Curve = c
	return nil
}

// SetBeating writes a whole note of the goal curve from one beating value; value in unit ("cent" or "hz").
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
	if err := c.SetBeating(tuning.Note(note), value, s.active.A4); err != nil {
		logger.Warn(logger.MsgSessionRejected, logger.Err(err))
		return ErrInvalidValue
	}
	s.active.Curve = c
	return nil
}

// SetAsymmetry divides the next beating either side of the reference reed, in percent, -100..100.
// It does not re-derive anchors already entered.
func (s *Service) SetAsymmetry(percent float64) error {
	if math.IsNaN(percent) || percent < -target.MaxAsymmetry || percent > target.MaxAsymmetry {
		logger.Warn(logger.MsgSessionRejected,
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

// SetInterpolate ramps between anchors (on) or steps to the nearer one (off).
func (s *Service) SetInterpolate(on bool) error {
	return s.setCurveFlag(func(c *target.Curve) { c.Interpolate = on })
}

// SetExtrapolateLeft holds the first anchor flat below it (on) or reads zero there (off).
func (s *Service) SetExtrapolateLeft(on bool) error {
	return s.setCurveFlag(func(c *target.Curve) { c.ExtrapolateLeft = on })
}

// SetExtrapolateRight is SetExtrapolateLeft above the last anchor.
func (s *Service) SetExtrapolateRight(on bool) error {
	return s.setCurveFlag(func(c *target.Curve) { c.ExtrapolateRight = on })
}

func (s *Service) setCurveFlag(set func(*target.Curve)) error {
	s.mu.Lock()
	defer s.unlockAndPublish()
	if s.active == nil {
		return ErrNoSession
	}
	c := s.editCurveLocked()
	set(c)
	s.active.Curve = c
	return nil
}

// SetRefReed names the reed the curve treats as being at pitch, or NoRefReed for none.
func (s *Service) SetRefReed(reed int) error {
	s.mu.Lock()
	defer s.unlockAndPublish()
	if s.active == nil {
		return ErrNoSession
	}

	c := s.editCurveLocked()
	if reed != target.NoRefReed && (reed < 0 || reed >= c.ReedCount) {
		return ErrInvalidReed
	}

	c.RefReed = reed
	s.active.Curve = c
	return nil
}

// SetCurveUnit chooses the unit the next anchors are typed in; storage stays cents.
func (s *Service) SetCurveUnit(unit string) error {
	u, err := parseUnit(unit)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.unlockAndPublish()
	if s.active == nil {
		return ErrNoSession
	}

	c := s.editCurveLocked()
	c.Unit = u
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
	logger.Info(logger.MsgSessionCurveDropped)
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
