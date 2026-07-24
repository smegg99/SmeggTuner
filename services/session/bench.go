// The pulled register is not in the audio, so it is said once and stamped onto every take. The
// register is also the reed count, so there is no separate control to get out of step with it.
package session

import (
	coresession "smegg.me/smeggtuner/core/session"
)

// BenchDTO is the bench as the toolbar draws it.
type BenchDTO struct {
	// Register is the switch that is pulled, by name; empty when the instrument has no registers.
	Register string `json:"register"`
	// Reeds is what the engine is asked to resolve: the pulled register's, or the instrument's own.
	Reeds int `json:"reeds"`
	// Banks is the pulled register's, in card order: the columns this take will land in.
	Banks []coresession.Bank `json:"banks"`
	// Bass says the bench faces the bass keyboard; the fields above describe the treble side and go
	// quiet while it does. Only an instrument with a declared bass machine can face it.
	Bass bool `json:"bass"`
	// BassRegister is the pulled bass switch, or empty for the whole machine (a fixed bass has no
	// switches at all). BassFeet is what sounds, largest foot first: the ladder being tuned.
	BassRegister string `json:"bassRegister"`
	BassFeet     []int  `json:"bassFeet"`
}

// SetRegister says which switch is pulled. It is the reed count too.
func (s *Service) SetRegister(name string) error {
	s.mu.Lock()
	if s.active == nil {
		s.mu.Unlock()
		return ErrNoSession
	}

	// Empty register is always allowed: not tracking banks.
	if name != "" {
		if _, ok := s.active.Instrument.Register(name); !ok {
			s.mu.Unlock()
			return ErrNoRegister
		}
	}
	s.register = name
	s.mu.Unlock()

	s.emitActive()
	return nil
}

// benchLocked builds the bench; caller holds s.mu.
func (s *Service) benchLocked() BenchDTO {
	if s.active == nil {
		return BenchDTO{}
	}

	if s.bassSide && s.active.Instrument.BassReeds > 0 {
		b := BenchDTO{Bass: true, BassRegister: s.bassRegister}
		b.BassFeet = coresession.BassFeet(s.active.Instrument.BassReeds)
		if r, ok := s.active.Instrument.BassRegister(s.bassRegister); ok {
			b.BassFeet = append([]int(nil), r.Feet...)
		}
		b.Reeds = len(b.BassFeet)
		return b
	}

	b := BenchDTO{
		Register: s.register,
		Reeds:    s.active.Instrument.ReedCount,
	}
	if r, ok := s.active.Instrument.Register(s.register); ok {
		b.Reeds = r.ReedCount()
		b.Banks = append([]coresession.Bank(nil), r.Banks...)
	}
	return b
}

// defaultRegisterLocked picks the register that sounds every reed, so a full sweep records the
// right columns without anyone choosing.
func (s *Service) defaultRegisterLocked() string {
	i := s.active.Instrument
	for _, r := range i.Registers {
		if r.ReedCount() == i.ReedCount {
			return r.Name
		}
	}
	if len(i.Registers) > 0 {
		return i.Registers[0].Name
	}
	return ""
}

// SetBass turns the bench toward the bass keyboard, or back; only an instrument with a declared
// bass machine has one to face.
func (s *Service) SetBass(on bool) error {
	s.mu.Lock()
	if s.active == nil {
		s.mu.Unlock()
		return ErrNoSession
	}
	if on && s.active.Instrument.BassReeds == 0 {
		s.mu.Unlock()
		return ErrNoBassMachine
	}
	s.bassSide = on
	s.mu.Unlock()

	s.emitActive()
	return nil
}

// SetBassRegister says which bass switch is pulled; empty is the whole machine.
func (s *Service) SetBassRegister(name string) error {
	s.mu.Lock()
	if s.active == nil {
		s.mu.Unlock()
		return ErrNoSession
	}
	if name != "" {
		if _, ok := s.active.Instrument.BassRegister(name); !ok {
			s.mu.Unlock()
			return ErrNoBassRegister
		}
	}
	s.bassRegister = name
	s.mu.Unlock()

	s.emitActive()
	return nil
}

// stampLocked stamps the bench onto a take, unless the take already names a register (import or
// replay).
func (s *Service) stampLocked(t *coresession.Take) {
	if t.Register != "" || t.Bass {
		return
	}
	if s.bassSide {
		t.Bass = true
		t.Register = s.bassRegister
		return
	}
	t.Register = s.register
}
