package session

import "fmt"

// maxTolerance is the widest judging window, in cents; matches the app config schema.
const maxTolerance = 50.0

// Has reports whether the instrument sounds this bank.
func (i Instrument) Has(b Bank) bool {
	for _, have := range i.Banks {
		if have == b {
			return true
		}
	}
	return false
}

// Register finds a register by name.
func (i Instrument) Register(name string) (Register, bool) {
	for _, r := range i.Registers {
		if r.Name == name {
			return r, true
		}
	}
	return Register{}, false
}

// ReedCount is how many reeds this register sounds (len(Banks)).
func (r Register) ReedCount() int { return len(r.Banks) }

// validate checks the instrument's internal consistency.
func (i Instrument) validate() error {
	// No banks named is a pre-banks instrument: it still loads and tunes, it just cannot print a per-bank card.
	if len(i.Banks) > 0 {
		if err := validBanks(i.Banks); err != nil {
			return err
		}
	}

	for _, r := range i.Registers {
		if err := validBanks(r.Banks); err != nil {
			return fmt.Errorf("register %q: %w", r.Name, err)
		}

		// A register may not sound a rank the instrument does not have.
		for _, b := range r.Banks {
			if len(i.Banks) > 0 && !i.Has(b) {
				return fmt.Errorf("register %q: %w: %q", r.Name, ErrBank, b)
			}
		}

		if err := validReeds(r.ReedCount()); err != nil {
			return fmt.Errorf("register %q: %w", r.Name, err)
		}
	}

	// A keyboard runs upward; zero means unset, which is allowed.
	if i.Lo != 0 || i.Hi != 0 {
		if !i.Lo.Valid() || !i.Hi.Valid() || i.Lo > i.Hi {
			return fmt.Errorf("%w: %d to %d", ErrRange, i.Lo, i.Hi)
		}
	}

	// Reference pitch when set; zero falls back to the app default.
	if i.A4 < 0 {
		return ErrA4
	}

	// Judging windows when set; zero means unset, otherwise positive and within the schema.
	for _, tol := range []float64{i.Tolerance, i.BeatTolerance} {
		if tol != 0 && (tol <= 0 || tol > maxTolerance) {
			return ErrTolerance
		}
	}
	return i.validateBass()
}

// Tolerances returns this instrument's windows, or the given defaults where unset; the one place instrument-over-default precedence lives.
func (i Instrument) Tolerances(defTol, defBeat float64) (tol, beat float64) {
	tol, beat = defTol, defBeat
	if i.Tolerance > 0 {
		tol = i.Tolerance
	}
	if i.BeatTolerance > 0 {
		beat = i.BeatTolerance
	}
	return tol, beat
}

// validTake refuses a take naming a register the instrument does not have; a bass take checks the
// bass switches, a treble take the treble ones.
func (i Instrument) validTake(t Take) error {
	if t.Register == "" {
		return nil
	}
	if t.Bass {
		if len(i.BassRegisters) > 0 {
			if _, ok := i.BassRegister(t.Register); !ok {
				return fmt.Errorf("%w: %q", ErrBassRegister, t.Register)
			}
		}
		return nil
	}
	if len(i.Registers) > 0 {
		if _, ok := i.Register(t.Register); !ok {
			return fmt.Errorf("%w: %q", ErrRegister, t.Register)
		}
	}
	return nil
}

// ValidInstrument validates an instrument for the service layer.
func ValidInstrument(i Instrument) error {
	if err := validReeds(i.ReedCount); err != nil {
		return err
	}
	return i.validate()
}

func validReeds(n int) error {
	if n < MinReeds || n > MaxReeds {
		return fmt.Errorf("%w: %d", ErrReedCount, n)
	}
	return nil
}
