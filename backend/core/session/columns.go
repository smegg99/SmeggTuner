package session

import (
	"slices"
	"sort"
)

// The card's columns are the instrument's own ranks, in card order, and they are the goal curve's
// columns too. A reading's slots are positional: slot i is the pulled register's i-th bank in card
// order, or on the bass side the i-th sounding foot largest first. Everything that scores a reading
// against the curve or lays it out in a table needs the same answer, so it is stated once here.

// TakeBanks is the ranks a treble take's switch sounds. Nil when the take names no switch, or one
// the instrument no longer has; the caller then numbers the slots rather than naming ranks it is
// not sure of.
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

// Columns is the card column each of a take's slots answers to: by bank on the treble side, by
// foot on the bass side. Nil when uncertain; the caller then falls back to slot positions.
func (i Instrument) Columns(t Take) []int {
	if t.Bass {
		return FootColumns(BassFeet(i.BassReeds), i.SoundingFeet(t.Register), len(t.Reeds))
	}
	return BankColumns(i.Banks, i.TakeBanks(t), len(t.Reeds))
}

// ColumnOf is the column slot answers to in cols (see Columns), or the slot's own position when cols is nil.
func ColumnOf(cols []int, slot int) int {
	if slot < len(cols) {
		return cols[slot]
	}
	return slot
}

// BankColumns maps a reading's slots onto the card's columns: each slot's bank (see SlotBanks)
// and that bank's place in the instrument's own list. Nil on any uncertain mapping; the caller
// then falls back to slot positions, which is the same answer when the register sounds every bank.
func BankColumns(card, register []Bank, slots int) []int {
	banks := SlotBanks(register, slots)
	if banks == nil || len(card) == 0 {
		return nil
	}
	cols := make([]int, len(banks))
	for n, b := range banks {
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
func FootColumns(card, sounding []int, slots int) []int {
	feet := SlotFeet(sounding, slots)
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

// SlotFeet is the foot each of a bass reading's slots measured: the sounding ladder, largest foot
// first. Nil when the ladder does not have exactly that many ranks.
func SlotFeet(sounding []int, slots int) []int {
	if slots == 0 || len(sounding) != slots {
		return nil
	}
	ladder := append([]int(nil), sounding...)
	sort.Sort(sort.Reverse(sort.IntSlice(ladder)))
	return ladder
}
