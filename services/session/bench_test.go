package session

import (
	"errors"
	"testing"

	coresession "smegg.me/smeggtuner/core/session"
)

func musette() coresession.Instrument {
	m := func(b ...coresession.Bank) []coresession.Bank { return b }
	return coresession.Instrument{
		ReedCount: 3,
		Banks:     m(coresession.BankM1, coresession.BankM2, coresession.BankM3),
		Registers: []coresession.Register{
			{Name: "MMM", Banks: m(coresession.BankM1, coresession.BankM2, coresession.BankM3)},
			{Name: "M1M2", Banks: m(coresession.BankM1, coresession.BankM2)},
			{Name: "M2", Banks: m(coresession.BankM2)},
		},
	}
}

func bench(t *testing.T, s *Service) BenchDTO {
	t.Helper()
	a := s.Active()
	if a == nil {
		t.Fatal("no session")
	}
	return a.Bench
}

func openMusette(t *testing.T) *Service {
	t.Helper()
	s := service(t)
	if _, err := s.Create(NewSessionDTO{Name: "Bench", Instrument: musette()}); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestASessionOpensOnTheRegisterThatSoundsEverything(t *testing.T) {
	b := bench(t, openMusette(t))
	if b.Register != "MMM" {
		t.Fatalf("register = %q, want the one that sounds all three", b.Register)
	}
	if b.Reeds != 3 {
		t.Fatalf("reeds = %d, want 3", b.Reeds)
	}
}

// The register is the reed count: pulling one tells the engine how many reeds to resolve.
func TestPullingARegisterTellsTheEngineWhatToResolve(t *testing.T) {
	s := openMusette(t)

	if err := s.SetRegister("M1M2"); err != nil {
		t.Fatal(err)
	}
	if got := s.Goal().Reeds; got != 2 {
		t.Fatalf("the engine is resolving %d reeds out of a two reed register", got)
	}

	if err := s.SetRegister("M2"); err != nil {
		t.Fatal(err)
	}
	if got := s.Goal().Reeds; got != 1 {
		t.Fatalf("the engine is resolving %d reeds out of a single rank", got)
	}
	if b := bench(t, s); len(b.Banks) != 1 || b.Banks[0] != coresession.BankM2 {
		t.Fatalf("the bench says this take lands in %v", b.Banks)
	}
}

func TestAnUnknownRegisterCannotBePulled(t *testing.T) {
	s := openMusette(t)
	if err := s.SetRegister("LMMM"); !errors.Is(err, ErrNoRegister) {
		t.Fatalf("set = %v, want %v", err, ErrNoRegister)
	}
	if bench(t, s).Register != "MMM" {
		t.Fatal("a refused register was pulled anyway")
	}
}

// Every take is stamped with the bench, because none of it is in the audio.
func TestEveryTakeIsStampedWithTheBench(t *testing.T) {
	s := openMusette(t)
	if err := s.SetRegister("M1M2"); err != nil {
		t.Fatal(err)
	}
	if err := s.UpsertTake(take(69, 442, -3, 0)); err != nil {
		t.Fatal(err)
	}

	snap := s.Snapshot()
	got := snap.Takes[0]
	if got.Register != "M1M2" {
		t.Fatalf("the take says it was played on %q", got.Register)
	}
}

// With no registers, takes name no register rather than inventing one that does not exist.
func TestASessionWithNoRegistersStillRecords(t *testing.T) {
	s := service(t)
	if _, err := s.Create(NewSessionDTO{Name: "Bench", Instrument: instrument(3)}); err != nil {
		t.Fatal(err)
	}

	b := bench(t, s)
	if b.Register != "" || b.Reeds != 3 {
		t.Fatalf("bench = %+v, want no register and the instrument's own reeds", b)
	}

	if err := s.UpsertTake(take(69, 442, -3, 0, 3)); err != nil {
		t.Fatal(err)
	}
	if got := s.Snapshot().Takes[0].Register; got != "" {
		t.Fatalf("the take named a register the instrument has not got: %q", got)
	}
}
