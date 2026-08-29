package session

import (
	"slices"
	"testing"
)

// Slot i is the register's i-th bank in card order, whatever order the switch was typed in.
func TestSlotsAreTheRegistersBanksInCardOrder(t *testing.T) {
	card := []Bank{BankL, BankM1, BankM2, BankH}
	if got := SlotBanks([]Bank{BankH, BankL}, 2); !slices.Equal(got, []Bank{BankL, BankH}) {
		t.Fatalf("SlotBanks(HL) = %v, want L then H", got)
	}
	if got := BankColumns(card, []Bank{BankH, BankL}, 2); !slices.Equal(got, []int{0, 3}) {
		t.Fatalf("an LH take lands in columns %v, want 0 and 3", got)
	}
	if got := BankColumns(card, []Bank{BankM1, BankM2}, 2); !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("an MM take lands in columns %v, want 1 and 2", got)
	}
}

// A reading the register cannot account for is numbered, never named by a guess.
func TestAnUncertainMappingIsNil(t *testing.T) {
	card := []Bank{BankL, BankM1, BankM2, BankH}
	for name, got := range map[string][]int{
		"more slots than banks": BankColumns(card, []Bank{BankM2}, 3),
		"fewer slots":           BankColumns(card, []Bank{BankL, BankM1, BankH}, 2),
		"a rank off the card":   BankColumns(card, []Bank{BankM3}, 1),
		"no card":               BankColumns(nil, []Bank{BankM1}, 1),
		"no register":           BankColumns(card, nil, 1),
	} {
		if got != nil {
			t.Errorf("%s: columns %v, want nil", name, got)
		}
	}
	if SlotBanks(nil, 0) != nil || SlotFeet(nil, 0) != nil {
		t.Error("nothing measured names nothing")
	}
}

// On the bass side slot i is the i-th sounding foot, largest first.
func TestBassSlotsAreTheSoundingFeetLargestFirst(t *testing.T) {
	card := BassFeet(5)
	if got := SlotFeet([]int{8, 16}, 2); !slices.Equal(got, []int{16, 8}) {
		t.Fatalf("SlotFeet = %v, want 16 then 8", got)
	}
	if got := FootColumns(card, []int{8, 16}, 2); !slices.Equal(got, []int{1, 2}) {
		t.Fatalf("a 16.8 take lands in columns %v, want 1 and 2", got)
	}
	if got := FootColumns(card, []int{16, 8}, 3); got != nil {
		t.Fatalf("three slots off a two-foot switch mapped to %v", got)
	}
}
