package session

import (
	"errors"
	"fmt"
	"math/bits"
	"sort"
	"strconv"
	"strings"

	"smegg.me/smeggtuner/backend/core/dsp"
)

// The bass machine: reed ranks stacked in pure octaves, four to six of them on most instruments,
// every button sounding its whole ladder. Ranks are named by feet the organ way, largest (lowest)
// first, always ending at 2': a five-voice machine is 32.16.8.4.2. Older machines are fixed - no
// switches, all ranks always on - newer ones declare BassRegisters. Tuning-wise a bass button is a
// compound register that starts near 40 Hz, and the engine measures it exactly like 16'+8'+4'.

// Bass machines sound 2..6 ranks; the common ones 4 or 5.
const (
	MinBassReeds = 2
	MaxBassReeds = 6
)

var (
	ErrBassReeds    = fmt.Errorf("session: bass reed count outside %d..%d", MinBassReeds, MaxBassReeds)
	ErrBassFoot     = errors.New("session: not a rank of this bass machine")
	ErrBassRegister = errors.New("session: the instrument has no such bass register")
)

// BassRegister is one of the bass machine's switches: which ranks it sounds, by foot.
type BassRegister struct {
	Name string `json:"name"`
	Feet []int  `json:"feet"`
}

// BassFeet is the machine's ranks by foot for a given voice count, largest first: 5 -> 32,16,8,4,2.
func BassFeet(count int) []int {
	if count < MinBassReeds || count > MaxBassReeds {
		return nil
	}
	feet := make([]int, count)
	for i := range feet {
		feet[i] = 1 << (count - i)
	}
	return feet
}

// BassRegister finds a bass register by name.
func (i Instrument) BassRegister(name string) (BassRegister, bool) {
	for _, r := range i.BassRegisters {
		if r.Name == name {
			return r, true
		}
	}
	return BassRegister{}, false
}

// OctavesOfFeet maps sounding bass ranks onto the engine's compound layout: the largest foot is the
// base, and each halving of the foot climbs an octave. Nil unless at least two ranks sound - one
// rank is the plain tuner's job.
func OctavesOfFeet(feet []int) []dsp.OctaveRequest {
	if len(feet) < 2 {
		return nil
	}
	sorted := append([]int(nil), feet...)
	sort.Sort(sort.Reverse(sort.IntSlice(sorted)))
	out := make([]dsp.OctaveRequest, len(sorted))
	for i, f := range sorted {
		out[i] = dsp.OctaveRequest{Offset: 12 * (bits.Len(uint(sorted[0])) - bits.Len(uint(f))), Reeds: 1}
	}
	return out
}

// SoundingFeet is the ladder a bass register sounds, largest foot first; the whole machine when no
// switch is named. Nil when the machine is undeclared or the switch is unknown.
func (i Instrument) SoundingFeet(register string) []int {
	machine := BassFeet(i.BassReeds)
	if machine == nil {
		return nil
	}
	if register == "" {
		return machine
	}
	r, ok := i.BassRegister(register)
	if !ok {
		return nil
	}
	feet := append([]int(nil), r.Feet...)
	sort.Sort(sort.Reverse(sort.IntSlice(feet)))
	return feet
}

// validBassFeet checks a rank selection against the declared machine.
func validBassFeet(feet []int, machine []int) error {
	if len(feet) == 0 {
		return fmt.Errorf("%w: a bass register must sound at least one rank", ErrBassFoot)
	}
	seen := map[int]bool{}
	for _, f := range feet {
		ok := false
		for _, m := range machine {
			if f == m {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("%w: %d'", ErrBassFoot, f)
		}
		if seen[f] {
			return fmt.Errorf("%w: %d' twice", ErrBassFoot, f)
		}
		seen[f] = true
	}
	return nil
}

// validateBass checks the instrument's bass section; no section at all is the ordinary old case.
func (i Instrument) validateBass() error {
	if i.BassReeds == 0 {
		if len(i.BassRegisters) > 0 {
			return fmt.Errorf("%w: registers without a declared machine", ErrBassReeds)
		}
		return nil
	}
	if i.BassReeds < MinBassReeds || i.BassReeds > MaxBassReeds {
		return fmt.Errorf("%w: %d", ErrBassReeds, i.BassReeds)
	}
	if len(i.BassRegisters) > MaxBassRegisters {
		return fmt.Errorf("%w: %d", ErrTooManyRegisters, len(i.BassRegisters))
	}

	machine := BassFeet(i.BassReeds)
	// Same rule as the treble side: a bass switch is the feet it sounds, and two of them sounding
	// the same feet are one switch written twice.
	seenName := make(map[string]bool, len(i.BassRegisters))
	seenFeet := make(map[string]bool, len(i.BassRegisters))

	for _, r := range i.BassRegisters {
		if len(r.Name) > MaxSwitchName {
			return fmt.Errorf("%w: %q", ErrSwitchName, r.Name)
		}
		if seenName[r.Name] {
			return fmt.Errorf("%w: %q", ErrDuplicateRegister, r.Name)
		}
		seenName[r.Name] = true

		if key := feetKey(r.Feet); seenFeet[key] {
			return fmt.Errorf("%w: %q sounds what another bass switch already does", ErrDuplicateRegister, r.Name)
		} else {
			seenFeet[key] = true
		}

		if err := validBassFeet(r.Feet, machine); err != nil {
			return fmt.Errorf("bass register %q: %w", r.Name, err)
		}
	}
	return nil
}

// feetKey identifies a bass switch by what it sounds, largest foot first, so the order it was
// clicked in is not part of its identity.
func feetKey(feet []int) string {
	ordered := append([]int(nil), feet...)
	sort.Sort(sort.Reverse(sort.IntSlice(ordered)))

	parts := make([]string, len(ordered))
	for i, f := range ordered {
		parts[i] = strconv.Itoa(f)
	}
	return strings.Join(parts, "+")
}
