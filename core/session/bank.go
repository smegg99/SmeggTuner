package session

import (
	"errors"
	"fmt"
	"strings"

	"smegg.me/smeggtuner/core/dsp"
)

// Bank is a rank of reeds.
type Bank string

// The banks, in card order.
const (
	BankL  Bank = "L"  // 16 foot, an octave below
	BankM1 Bank = "M1" // 8 foot, at pitch
	BankM2 Bank = "M2" // 8 foot, sharp of M1 - the musette beat
	BankM3 Bank = "M3" // 8 foot, flat of M1
	BankM4 Bank = "M4" // 8 foot, a fourth musette rank
	BankH  Bank = "H"  // 4 foot, an octave above
)

// Banks is every bank, in card order.
var Banks = []Bank{BankL, BankM1, BankM2, BankM3, BankM4, BankH}

var (
	ErrBank      = errors.New("session: not a reed bank")
	ErrBanks     = errors.New("session: a register must sound at least one bank")
	ErrBankTwice = errors.New("session: the same bank twice")
	ErrRange     = errors.New("session: the keyboard runs low to high")
	ErrRegister  = errors.New("session: the instrument has no such register")
	ErrNoBanks   = errors.New("session: the instrument sounds no reed banks")
	ErrTolerance = errors.New("session: a tolerance is a positive number of cents")
)

func (b Bank) Valid() bool {
	switch b {
	case BankL, BankM1, BankM2, BankM3, BankM4, BankH:
		return true
	}
	return false
}

// Octave is where the bank sounds, in semitones from the played key: L an octave under, H one over.
func (b Bank) Octave() int {
	switch b {
	case BankL:
		return -12
	case BankH:
		return 12
	}
	return 0
}

// AssignBanks names the bank each of a take's reeds landed in: a reed claims the register's first
// unclaimed bank in its own octave, in card order. A take from before octaves (every reed at zero)
// claims positionally when the counts match - the old contract. Nil when any claim fails; the
// caller then numbers the reeds instead of naming ranks it is not sure of.
func AssignBanks(register []Bank, reeds []dsp.ReedMeasure) []Bank {
	if len(register) == 0 || len(reeds) == 0 {
		return nil
	}
	octaved := false
	for _, r := range reeds {
		if r.Octave != 0 {
			octaved = true
			break
		}
	}
	if !octaved {
		if len(reeds) != len(register) {
			return nil
		}
		return append([]Bank(nil), register...)
	}
	used := make([]bool, len(register))
	out := make([]Bank, len(reeds))
	for i, r := range reeds {
		claimed := false
		for j, b := range register {
			if !used[j] && b.Octave() == r.Octave {
				used[j], out[i], claimed = true, b, true
				break
			}
		}
		if !claimed {
			return nil
		}
	}
	return out
}

// OctavesOf maps a register's banks onto the engine's compound layout: one band per octave the
// register sounds, ascending, each carrying its rank count. Nil for a register that stays in the
// key's own octave - the engine's single band (and its musette machinery) is the right tool there.
func OctavesOf(banks []Bank) []dsp.OctaveRequest {
	counts := map[int]int{}
	spans := false
	for _, b := range banks {
		off := b.Octave()
		counts[off]++
		if off != 0 {
			spans = true
		}
	}
	if !spans {
		return nil
	}
	var out []dsp.OctaveRequest
	for _, off := range []int{-12, 0, 12} {
		if counts[off] > 0 {
			out = append(out, dsp.OctaveRequest{Offset: off, Reeds: counts[off]})
		}
	}
	return out
}

// ParseBanks reads a register name (LMMM, MMM, M2) as the banks it sounds.
func ParseBanks(name string) ([]Bank, error) {
	s := strings.ToUpper(strings.TrimSpace(name))
	if s == "" {
		return nil, fmt.Errorf("%w: %q", ErrBanks, name)
	}

	var out []Bank
	ms := 0

	for i := 0; i < len(s); i++ {
		switch s[i] {
		case 'L':
			out = append(out, BankL)

		case 'H':
			out = append(out, BankH)

		case 'M':
			// M1..M4 written out, or a run of bare Ms counted off in order.
			if i+1 < len(s) && s[i+1] >= '1' && s[i+1] <= '4' {
				out = append(out, Bank("M"+string(s[i+1])))
				i++
				break
			}
			ms++
			if ms > 4 {
				return nil, fmt.Errorf("%w: %q sounds more than four M ranks", ErrBank, name)
			}
			out = append(out, Bank(fmt.Sprintf("M%d", ms)))

		default:
			return nil, fmt.Errorf("%w: %q", ErrBank, name)
		}
	}

	return out, validBanks(out)
}

func validBanks(banks []Bank) error {
	if len(banks) == 0 {
		return ErrBanks
	}

	seen := make(map[Bank]bool, len(banks))
	for _, b := range banks {
		if !b.Valid() {
			return fmt.Errorf("%w: %q", ErrBank, b)
		}
		if seen[b] {
			return fmt.Errorf("%w: %q", ErrBankTwice, b)
		}
		seen[b] = true
	}
	return nil
}
