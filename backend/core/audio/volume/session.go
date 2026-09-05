// backend/core/audio/volume/session.go
package volume

import "errors"

// Endpoint is one capture device's level and mute in the OS mixer, the level on the mixer's own
// 0..1 scale (PulseAudio may report more than 1).
type Endpoint interface {
	Level() (float64, error)
	SetLevel(float64) error
	Muted() (bool, error)
	SetMuted(bool) error
}

// Session steers one capture's endpoint. Begin keeps what the OS had and applies the device's own
// level; Restore hands the OS its level and mute back.
type Session struct {
	ep      Endpoint
	ctl     Control
	osLevel float64
	osMuted bool
}

// Begin saves the endpoint's level and mute, then applies stored, where this device was last left.
func Begin(ep Endpoint, stored float64) (*Session, error) {
	level, err := ep.Level()
	if err != nil {
		return nil, err
	}
	muted, err := ep.Muted()
	if err != nil {
		return nil, err
	}
	if err := ep.SetLevel(min(maxLevel, max(0, stored))); err != nil {
		return nil, err
	}
	return &Session{ep: ep, osLevel: level, osMuted: muted}, nil
}

// Interval runs the control law on one level interval. A mute it cannot read counts as unmuted and
// a level it cannot read as full, as in the reference.
func (s *Session) Interval(iv Interval) error {
	muted, err := s.ep.Muted()
	muted = muted && err == nil
	level := float32(1)
	if !muted {
		if v, err := s.ep.Level(); err == nil {
			level = float32(min(1, max(0, v)))
		}
	}
	next, set := s.ctl.Step(iv, muted, level)
	if !set {
		return nil
	}
	return s.ep.SetLevel(float64(next))
}

// Left is the level the capture leaves the device at, for the device's store.
func (s *Session) Left() (float64, error) {
	level, err := s.ep.Level()
	return min(maxLevel, max(0, level)), err
}

// Restore hands the OS the level and mute Begin found.
func (s *Session) Restore() error {
	return errors.Join(s.ep.SetLevel(s.osLevel), s.ep.SetMuted(s.osMuted))
}
