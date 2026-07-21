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

// stampLocked stamps the bench onto a take, unless the take already names a register (import or
// replay).
func (s *Service) stampLocked(t *coresession.Take) {
	if t.Register == "" {
		t.Register = s.register
	}
}
