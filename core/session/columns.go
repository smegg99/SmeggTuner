// core/session/columns.go
package session

import (
	"slices"
	"sort"

	"smegg.me/smeggtuner/core/dsp"
)

// The card's columns are the instrument's own ranks, in card order, and they are the goal curve's
// columns too. A reading names its ranks by octave - AssignBanks on the treble side, the foot
// ladder on the bass - and this is where those land. Everything that scores a reading against the
// curve or lays it out in a table needs the same answer, so it is stated once here.

// TakeBanks is the ranks a take's switch sounds, mirroring TakeFeet on the bass side. Nil when the
// take names no switch, or one the instrument no longer has - the caller then numbers the reeds
// rather than naming ranks it is not sure of.
func (i Instrument) TakeBanks(t Take) []Bank {
	if t.Register == "" || t.Bass {
		return nil
	}
	r, ok := i.Register(t.Register)
	if !ok {
		return nil
	}
	return r.Banks
}

// BankColumns maps a reading's reeds onto the card's columns: each reed claims one of the pulled
// register's banks in its own octave, and that bank's place in the instrument's own list is the
// column. Nil on any uncertain mapping; the caller then falls back to the reeds' own positions,
// which is the same answer only when the register sounds every bank.
func BankColumns(card, register []Bank, reeds []dsp.ReedMeasure) []int {
	assigned := AssignBanks(register, reeds)
	if assigned == nil || len(card) == 0 {
		return nil
	}
	cols := make([]int, len(assigned))
	for n, b := range assigned {
		col := slices.Index(card, b)
		if col < 0 {
			return nil // a rank the instrument does not have
		}
		cols[n] = col
	}
	return cols
}

// FootColumns is the same on the bass side: sounding is what the pulled bass register sounds,
// card the machine's whole ladder, both largest foot first.
func FootColumns(card, sounding []int, reeds []dsp.ReedMeasure) []int {
	feet := AssignFeet(sounding, reeds)
	if feet == nil || len(card) == 0 {
		return nil
	}
	cols := make([]int, len(feet))
	for n, f := range feet {
		col := slices.Index(card, f)
		if col < 0 {
			return nil
		}
		cols[n] = col
	}
	return cols
}

// AssignFeet names the foot each of a reading's reeds landed in, mirroring AssignBanks: a reed's
// octave counts up from the sounding ladder's largest foot. A reading from before octaves (every
// reed at zero) claims positionally when the counts match - the old contract. Nil when any claim
// fails; the caller then numbers the reeds rather than naming ranks it is not sure of.
func AssignFeet(sounding []int, reeds []dsp.ReedMeasure) []int {
	if len(sounding) == 0 || len(reeds) == 0 {
		return nil
	}
	ladder := append([]int(nil), sounding...)
	sort.Sort(sort.Reverse(sort.IntSlice(ladder)))

	octaved := false
	for _, r := range reeds {
		if r.Octave != 0 {
			octaved = true
			break
		}
	}
	if !octaved {
		if len(reeds) != len(ladder) {
			return nil
		}
		return ladder
	}

	out := make([]int, len(reeds))
	for n, r := range reeds {
		idx := r.Octave / 12
		if r.Octave < 0 || r.Octave%12 != 0 || idx >= len(ladder) {
			return nil
		}
		out[n] = ladder[idx]
	}
	return out
}
