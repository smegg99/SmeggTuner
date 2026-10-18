package session

import (
	"smegg.me/smeggtuner/backend/core/target"
)

// SetInterpolate ramps between anchors (on) or steps to the nearer one (off).
func (s *Service) SetInterpolate(on bool) error {
	return s.setCurveFlag(func(c *target.Curve) { c.Interpolate = on })
}

// SetExtrapolateLeft continues the line through the first two anchors below the first, stopping at
// zero (on), or sets no goal there (off, the reference beat list).
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

// SetCurveUnit chooses the unit the next anchors are typed in; storage stays reference units.
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
