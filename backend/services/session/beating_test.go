package session

import (
	"errors"
	"math"
	"testing"

	"smegg.me/smeggtuner/backend/core/target"
)

func TestSetBeatingDerivesTheReeds(t *testing.T) {
	s := service(t)
	create(t, s, "Morino", 3, 440)

	if err := s.SetRefReed(1); err != nil {
		t.Fatal(err)
	}
	if err := s.SetBeating(69, 2, "hz"); err != nil {
		t.Fatal(err)
	}
	goal := s.Active().Curve.At(69)
	// The typed number is the beat between neighbouring reeds, read against A4 = 440 Hz.
	want := []float64{target.UnitHz.ToGoal(-2, 69), 0, target.UnitHz.ToGoal(2, 69)}
	for r := range want {
		if math.Abs(goal[r]-want[r]) > 1e-9 {
			t.Fatalf("reed %d = %v units, want %v: 2 Hz with reed 2 on zero is -2 / 0 / +2",
				r+1, goal[r], want[r])
		}
	}

	// The reference reed decides which one sits at pitch; moving it moves the tremolo, not its width.
	if err := s.SetRefReed(0); err != nil {
		t.Fatal(err)
	}
	if err := s.SetBeating(69, 2, "hz"); err != nil {
		t.Fatal(err)
	}
	c := s.Active().Curve
	if got := c.At(69)[0]; got != 0 {
		t.Fatalf("reed 1 = %v, want 0 with reed 1 on zero", got)
	}
	if got := c.Beating(69); math.Abs(got-2) > 1e-9 {
		t.Fatalf("beating = %v Hz, want the 2 that was typed", got)
	}

	if err := s.SetBeating(300, 2, "hz"); !errors.Is(err, ErrInvalidNote) {
		t.Fatalf("note outside the keyboard: err = %v, want %s", err, ErrInvalidNote.Key)
	}
	if err := s.SetBeating(69, 2, "furlong"); !errors.Is(err, ErrInvalidUnit) {
		t.Fatalf("unknown unit: err = %v, want %s", err, ErrInvalidUnit.Key)
	}
}

// A one-reed instrument cannot beat, and the value is refused rather than written as a zero.
func TestSetBeatingNeedsTwoReeds(t *testing.T) {
	s := service(t)
	create(t, s, "Bandoneon bass", 1, 440)

	if err := s.SetBeating(69, 2, "hz"); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("beating on one reed: err = %v, want %s", err, ErrInvalidValue.Key)
	}
	if c := s.Active().Curve; c != nil && len(c.Anchors) != 0 {
		t.Fatalf("a refused beating anchored something: %+v", c.Anchors)
	}
}

// Asymmetry is the reference's old rule: beat 2-3 is 1 + percent/100 times beat 1-2.
func TestSetAsymmetry(t *testing.T) {
	s := service(t)
	create(t, s, "Morino", 3, 440)

	if err := s.SetRefReed(1); err != nil {
		t.Fatal(err)
	}
	if err := s.SetAsymmetry(20); err != nil {
		t.Fatal(err)
	}
	if err := s.SetBeating(69, 2, "hz"); err != nil {
		t.Fatal(err)
	}
	c := s.Active().Curve
	if c.Asymmetry != 20 {
		t.Fatalf("asymmetry = %v, want 20", c.Asymmetry)
	}
	if got, want := c.At(69)[2], target.UnitHz.ToGoal(2.4, 69); math.Abs(got-want) > 1e-9 {
		t.Fatalf("reed 3 = %v, want %v: beat 2-3 is 1.2 times the 2 Hz of beat 1-2", got, want)
	}
	if got := c.Beating(69); math.Abs(got-2) > 1e-9 {
		t.Fatalf("beating = %v Hz, want the 2 that was typed", got)
	}

	if err := s.SetAsymmetry(101); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("asymmetry past 100 percent: err = %v, want %s", err, ErrInvalidValue.Key)
	}
	if err := s.SetAsymmetry(math.NaN()); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("NaN asymmetry: err = %v, want %s", err, ErrInvalidValue.Key)
	}
	if s.Active().Curve.Asymmetry != 20 {
		t.Fatal("a rejected asymmetry changed the curve")
	}
}
