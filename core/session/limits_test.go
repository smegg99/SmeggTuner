package session

import (
	"errors"
	"strings"
	"testing"
)

// A switch is the ranks it sounds. Two switches sounding the same ranks are the same switch under
// two names, and a card cannot show them apart - it has one column per rank, not per name. Nothing
// stopped it: a bass machine could carry five switches all sounding 32', each named differently.
func TestASwitchCannotBeAddedTwiceUnderAnotherName(t *testing.T) {
	i := Instrument{
		ReedCount: 2,
		Banks:     []Bank{BankL, BankM1},
		Registers: []Register{
			{Name: "LM", Banks: []Bank{BankL, BankM1}},
			{Name: "the same thing", Banks: []Bank{BankM1, BankL}}, // order is not identity
		},
	}
	if err := i.validate(); !errors.Is(err, ErrDuplicateRegister) {
		t.Errorf("two switches over the same ranks = %v, want ErrDuplicateRegister", err)
	}
}

func TestASwitchNameIsUsedOnce(t *testing.T) {
	i := Instrument{
		ReedCount: 2,
		Banks:     []Bank{BankL, BankM1},
		Registers: []Register{
			{Name: "LM", Banks: []Bank{BankL, BankM1}},
			{Name: "LM", Banks: []Bank{BankL}},
		},
	}
	if err := i.validate(); !errors.Is(err, ErrDuplicateRegister) {
		t.Errorf("two switches under one name = %v, want ErrDuplicateRegister", err)
	}
}

func TestABassSwitchCannotBeAddedTwiceUnderAnotherName(t *testing.T) {
	i := Instrument{
		ReedCount: 1,
		Banks:     []Bank{BankM1},
		BassReeds: 5,
		BassRegisters: []BassRegister{
			{Name: "one", Feet: []int{32}},
			{Name: "two", Feet: []int{32}},
		},
	}
	if err := i.validate(); !errors.Is(err, ErrDuplicateRegister) {
		t.Errorf("two bass switches over the same feet = %v, want ErrDuplicateRegister", err)
	}
}

// The widest treble panel in production is thirteen switches and the widest bass machine seven.
// The caps sit above both: they are there to stop a list nobody could read, not to argue with an
// instrument.
func TestTheSwitchListsAreBounded(t *testing.T) {
	t.Run("treble", func(t *testing.T) {
		i := Instrument{ReedCount: 1, Banks: []Bank{BankM1}}
		for n := range MaxRegisters + 1 {
			i.Registers = append(i.Registers, Register{
				Name: strings.Repeat("x", n+1), Banks: []Bank{BankM1},
			})
		}
		if err := i.validate(); !errors.Is(err, ErrTooManyRegisters) {
			t.Errorf("%d switches = %v, want ErrTooManyRegisters", len(i.Registers), err)
		}
	})

	t.Run("bass", func(t *testing.T) {
		i := Instrument{ReedCount: 1, Banks: []Bank{BankM1}, BassReeds: 6}
		feet := BassFeet(6)
		for n := range MaxBassRegisters + 1 {
			i.BassRegisters = append(i.BassRegisters, BassRegister{
				Name: strings.Repeat("x", n+1), Feet: []int{feet[n%len(feet)]},
			})
		}
		if err := i.validate(); err == nil {
			t.Errorf("%d bass switches were accepted", len(i.BassRegisters))
		}
	})
}

// A name is printed on the card, so it has to fit on one.
func TestASwitchNameIsBounded(t *testing.T) {
	i := Instrument{
		ReedCount: 1,
		Banks:     []Bank{BankM1},
		Registers: []Register{{Name: strings.Repeat("x", MaxSwitchName+1), Banks: []Bank{BankM1}}},
	}
	if err := i.validate(); !errors.Is(err, ErrSwitchName) {
		t.Errorf("an over-long switch name = %v, want ErrSwitchName", err)
	}

	ok := Instrument{
		ReedCount: 1,
		Banks:     []Bank{BankM1},
		Registers: []Register{{Name: strings.Repeat("x", MaxSwitchName), Banks: []Bank{BankM1}}},
	}
	if err := ok.validate(); err != nil {
		t.Errorf("a name of exactly the cap was refused: %v", err)
	}
}

// What a real instrument carries still has to pass.
func TestAnOrdinaryPanelIsFine(t *testing.T) {
	i := Instrument{
		ReedCount: 4,
		Banks:     []Bank{BankL, BankM1, BankM2, BankH},
		Registers: []Register{
			{Name: "LMMH", Banks: []Bank{BankL, BankM1, BankM2, BankH}},
			{Name: "LMM", Banks: []Bank{BankL, BankM1, BankM2}},
			{Name: "LM", Banks: []Bank{BankL, BankM1}},
			{Name: "MM", Banks: []Bank{BankM1, BankM2}},
			{Name: "M", Banks: []Bank{BankM1}},
			{Name: "M2", Banks: []Bank{BankM2}},
			{Name: "L", Banks: []Bank{BankL}},
			{Name: "H", Banks: []Bank{BankH}},
			{Name: "MH", Banks: []Bank{BankM1, BankH}},
			{Name: "LH", Banks: []Bank{BankL, BankH}},
			{Name: "LMH", Banks: []Bank{BankL, BankM1, BankH}},
		},
		BassReeds: 5,
		BassRegisters: []BassRegister{
			{Name: "Bass", Feet: []int{32, 16, 8, 4, 2}},
			{Name: "Soft", Feet: []int{16, 8}},
			{Name: "Tenor", Feet: []int{8, 4}},
		},
	}
	if err := i.validate(); err != nil {
		t.Errorf("an ordinary eleven-switch panel was refused: %v", err)
	}
}
