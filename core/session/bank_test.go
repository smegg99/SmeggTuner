package session

import (
	"errors"
	"slices"
	"testing"

	"smegg.me/smeggtuner/core/tuning"
)

func TestARegistersNameSaysWhatItSounds(t *testing.T) {
	cases := map[string][]Bank{
		"LMMM": {BankL, BankM1, BankM2, BankM3},
		"MMM":  {BankM1, BankM2, BankM3},
		"LM":   {BankL, BankM1},
		"M":    {BankM1},
		"LMH":  {BankL, BankM1, BankH},

		"M2":   {BankM2},
		"LM2":  {BankL, BankM2},
		"M1M3": {BankM1, BankM3},

		" lmmm ": {BankL, BankM1, BankM2, BankM3},
	}

	for name, want := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := ParseBanks(name)
			if err != nil {
				t.Fatalf("ParseBanks(%q): %v", name, err)
			}
			if !slices.Equal(got, want) {
				t.Fatalf("ParseBanks(%q) = %v, want %v", name, got, want)
			}
		})
	}
}

func TestANameThatIsNotTheNotationIsRefused(t *testing.T) {
	for _, name := range []string{"", "Musette", "Bandoneon", "Master", "LMMMMM", "M5", "L2"} {
		if got, err := ParseBanks(name); err == nil {
			t.Fatalf("ParseBanks(%q) = %v, want an error rather than a guess", name, got)
		}
	}
}

func TestAFourthMRankIsNotation(t *testing.T) {
	got, err := ParseBanks("LMMMM")
	if err != nil {
		t.Fatalf("ParseBanks(LMMMM) = %v, want the four-rank musette", err)
	}
	if want := []Bank{BankL, BankM1, BankM2, BankM3, BankM4}; !slices.Equal(got, want) {
		t.Fatalf("LMMMM = %v, want %v", got, want)
	}
	if got, err := ParseBanks("M4"); err != nil || !slices.Equal(got, []Bank{BankM4}) {
		t.Fatalf("ParseBanks(M4) = %v, %v", got, err)
	}
}

func TestARegisterCannotSoundOneRankTwice(t *testing.T) {
	if _, err := ParseBanks("M1M1"); !errors.Is(err, ErrBankTwice) {
		t.Fatalf("M1M1 = %v, want %v", err, ErrBankTwice)
	}
	i := Instrument{Banks: []Bank{BankM1, BankM1}}
	if err := i.validate(); !errors.Is(err, ErrBankTwice) {
		t.Fatalf("validate = %v, want %v", err, ErrBankTwice)
	}
}

func TestAReedCountIsJustHowManyBanksSound(t *testing.T) {
	r := Register{Name: "LMMM", Banks: []Bank{BankL, BankM1, BankM2, BankM3}}
	if r.ReedCount() != 4 {
		t.Fatalf("ReedCount = %d, want 4", r.ReedCount())
	}
	if (Register{Name: "M"}).ReedCount() != 0 {
		t.Fatal("a register that sounds nothing has no reeds")
	}
}

func TestARegisterCannotReachPastTheInstrument(t *testing.T) {
	i := Instrument{
		Banks:     []Bank{BankM1, BankM2},
		Registers: []Register{{Name: "Musette", Banks: []Bank{BankM1, BankM3}}},
	}
	if err := i.validate(); !errors.Is(err, ErrBank) {
		t.Fatalf("validate = %v, want %v", err, ErrBank)
	}
}

func TestATakeCannotNameARegisterTheInstrumentHasNotGot(t *testing.T) {
	i := Instrument{
		ReedCount: 3,
		Banks:     []Bank{BankM1, BankM2, BankM3},
		Registers: []Register{{Name: "MMM", Banks: []Bank{BankM1, BankM2, BankM3}}},
	}
	if err := i.validTake(Take{Note: 60, Register: "LMMM"}); !errors.Is(err, ErrRegister) {
		t.Fatalf("validTake = %v, want %v", err, ErrRegister)
	}
	if err := i.validTake(Take{Note: 60, Register: "MMM"}); err != nil {
		t.Fatalf("a register the instrument has was refused: %v", err)
	}

	// A take from before there were registers names none, and still loads.
	if err := i.validTake(Take{Note: 60}); err != nil {
		t.Fatalf("a take with no register was refused: %v", err)
	}
}

func TestTheKeyboardRunsUpward(t *testing.T) {
	if err := (Instrument{Lo: tuning.MaxNote, Hi: tuning.MinNote}).validate(); !errors.Is(err, ErrRange) {
		t.Fatal("a keyboard that runs backwards was accepted")
	}
	if err := (Instrument{}).validate(); err != nil {
		t.Fatalf("an instrument whose keyboard nobody has named was refused: %v", err)
	}
	if err := (Instrument{Lo: 60, Hi: 72}).validate(); err != nil {
		t.Fatalf("a keyboard was refused: %v", err)
	}
}

func TestAnInstrumentIsJudgedByItsOwnTolerances(t *testing.T) {
	tight := Instrument{Tolerance: 0.5, BeatTolerance: 1.0}
	if tol, beat := tight.Tolerances(3, 4); tol != 0.5 || beat != 1.0 {
		t.Fatalf("tol, beat = %v, %v, want the instrument's 0.5 and 1.0", tol, beat)
	}

	none := Instrument{}
	if tol, beat := none.Tolerances(3, 4); tol != 3 || beat != 4 {
		t.Fatalf("tol, beat = %v, %v, want the defaults 3 and 4", tol, beat)
	}

	half := Instrument{Tolerance: 0.8}
	if tol, beat := half.Tolerances(3, 4); tol != 0.8 || beat != 4 {
		t.Fatalf("tol, beat = %v, %v, want 0.8 and the default 4", tol, beat)
	}
}

func TestATolerancesRange(t *testing.T) {
	if err := (Instrument{ReedCount: 1, Tolerance: -1}).validate(); !errors.Is(err, ErrTolerance) {
		t.Fatalf("validate = %v, want %v", err, ErrTolerance)
	}
	if err := (Instrument{ReedCount: 1, BeatTolerance: 999}).validate(); !errors.Is(err, ErrTolerance) {
		t.Fatalf("validate = %v, want %v", err, ErrTolerance)
	}
	if err := (Instrument{ReedCount: 1, Tolerance: 2.5}).validate(); err != nil {
		t.Fatalf("a valid tolerance was refused: %v", err)
	}
}
