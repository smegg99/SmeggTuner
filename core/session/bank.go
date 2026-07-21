package session

import (
	"errors"
	"fmt"
	"strings"
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
