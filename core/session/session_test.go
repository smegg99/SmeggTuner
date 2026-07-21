package session

import (
	"errors"
	"testing"
	"time"

	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/target"
	"smegg.me/smeggtuner/core/tuning"
)

func reeds(n int) []dsp.ReedMeasure {
	out := make([]dsp.ReedMeasure, n)
	for i := range out {
		out[i] = dsp.ReedMeasure{Freq: 440 + float64(i), DevCents: float64(i)}
	}
	return out
}

func newTestSession(t *testing.T, reedCount int) *Session {
	t.Helper()
	s := New("Hohner Morino", Instrument{Make: "Hohner", Model: "Morino", ReedCount: reedCount}, 442)
	if err := s.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	return s
}

func take(n tuning.Note, reedCount int, at time.Time) Take {
	return Take{Note: n, At: at, Reeds: reeds(reedCount)}
}

func TestValidate(t *testing.T) {
	cases := map[string]struct {
		mut  func(*Session)
		want error
	}{
		"zero reeds":         {func(s *Session) { s.Instrument.ReedCount = 0 }, ErrReedCount},
		"nine reeds":         {func(s *Session) { s.Instrument.ReedCount = 9 }, ErrReedCount},
		"bank-less register": {func(s *Session) { s.Instrument.Registers = []Register{{Name: "Musette"}} }, ErrBanks},
		"register sounds a bank the instrument has not got": {func(s *Session) {
			s.Instrument.Banks = []Bank{BankM1, BankM2}
			s.Instrument.Registers = []Register{{Name: "Musette", Banks: []Bank{BankM1, BankM3}}}
		}, ErrBank},
		"the same bank twice": {func(s *Session) {
			s.Instrument.Banks = []Bank{BankM1, BankM1}
		}, ErrBankTwice},
		"a keyboard that runs backwards": {func(s *Session) {
			s.Instrument.Lo, s.Instrument.Hi = tuning.MaxNote, tuning.MinNote
		}, ErrRange},
		"a register the instrument has not got": {func(s *Session) {
			s.Instrument.Registers = []Register{{Name: "Musette", Banks: []Bank{BankM1, BankM2, BankM3}}}
			s.UpsertTake(Take{Note: tuning.MinNote, Reeds: reeds(3), Register: "Bandoneon"})
		}, ErrRegister},
		"zero a4":         {func(s *Session) { s.A4 = 0 }, ErrA4},
		"no id":           {func(s *Session) { s.ID = "" }, ErrBadID},
		"id with a slash": {func(s *Session) { s.ID = "../etc/passwd" }, ErrBadID},
		"note out of range": {func(s *Session) {
			s.UpsertTake(Take{Note: tuning.MaxNote + 1, Reeds: reeds(3)})
		}, ErrNote},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			s := newTestSession(t, 3)
			c.mut(s)
			err := s.Validate()
			if !errors.Is(err, c.want) {
				t.Fatalf("validate = %v, want %v", err, c.want)
			}
		})
	}

	ok := newTestSession(t, 5)
	ok.Instrument.Banks = []Bank{BankL, BankM1, BankM2, BankM3}
	ok.Instrument.Registers = []Register{
		{Name: "Musette", Banks: []Bank{BankM1, BankM2, BankM3}},
		{Name: "Bassoon", Banks: []Bank{BankL}},
	}
	ok.UpsertTake(take(tuning.MinNote, 5, time.Now()))
	ok.UpsertTake(take(tuning.MaxNote, 5, time.Now()))
	if err := ok.Validate(); err != nil {
		t.Fatalf("valid session rejected: %v", err)
	}
}

// A nil curve means "no goal yet": a first-class state, never an error. A broken curve is different.
func TestCurveValidation(t *testing.T) {
	s := newTestSession(t, 3)
	if s.Curve != nil {
		t.Fatal("a new session must start with no goal")
	}
	if err := s.Validate(); err != nil {
		t.Fatalf("a session with no goal must be valid: %v", err)
	}

	s.Curve = target.NewCurve("musette", 3)
	if err := s.Validate(); err != nil {
		t.Fatalf("a session with a goal must be valid: %v", err)
	}

	s.Curve.ReedCount = 0
	if err := s.Validate(); err == nil {
		t.Fatal("a broken curve passed validation")
	}
}
