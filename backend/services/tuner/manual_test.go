package tuner

import (
	"math"
	"testing"

	appconfig "smegg.me/smeggtuner/backend/common/config"
	"smegg.me/smeggtuner/backend/core/dsp"
)

// Stepping from automatic pins A4; semitone steps stay within E0..C9.
func TestStepping(t *testing.T) {
	s := freshService(t)
	steps := []struct {
		step func(int) (int, error)
		by   int
		want int
	}{
		{s.StepNote, 1, 69}, {s.StepNote, 1, 70}, {s.StepNote, -1, 69}, {s.StepNote, -1, 68},
	}
	for i, st := range steps {
		got, err := st.step(st.by)
		mustOK(t, err)
		if got != st.want || int(s.snapshot().ManualNote) != st.want {
			t.Fatalf("step %d = %d (engine %d), want %d", i, got, s.snapshot().ManualNote, st.want)
		}
	}
	for range 60 {
		_, _ = s.StepNote(1)
	}
	if got, _ := s.StepNote(1); got != 120 {
		t.Fatalf("top = %d, want C9 (120)", got)
	}
	for range 110 {
		_, _ = s.StepNote(-1)
	}
	if got, _ := s.StepNote(-1); got != 16 {
		t.Fatalf("bottom = %d, want E0 (16)", got)
	}
	_, err := s.StepNote(2)
	wantErr(t, err, ErrInvalidNote)
}

// Leaving automatic pins the note on screen, or A4 when none is shown.
func TestEnteringManual(t *testing.T) {
	s := freshService(t)
	if got, err := s.SetAuto(false); err != nil || got != 69 {
		t.Fatalf("manual with nothing shown = %d, %v, want A4 (69)", got, err)
	}
	if got, _ := s.SetAuto(true); got != 0 || s.snapshot().ManualNote != 0 {
		t.Fatalf("auto = %d, want 0", got)
	}
	s.noteShown(dsp.Measurement{Frame: dsp.FrameMeasured, Note: 52})
	s.noteShown(dsp.Measurement{Frame: dsp.FrameNone, Note: 99})
	if got, _ := s.SetAuto(false); got != 52 {
		t.Fatalf("manual = %d, want the shown 52", got)
	}
	if st := s.Settings(); st.ManualNote != 52 {
		t.Fatalf("settings manual note = %d, want 52", st.ManualNote)
	}
}

// A4 moves on a 0.1 Hz grid and refuses while locked.
func TestA4Control(t *testing.T) {
	s := freshService(t)
	mustOK(t, s.SetA4(441.26))
	if got := s.snapshot().A4; got != 441.3 {
		t.Fatalf("A4 = %v, want 441.3", got)
	}
	mustOK(t, s.SetA4(440))
	if got := appconfig.Get().Tuner.A4; got != 440 {
		t.Fatalf("config A4 = %v, want 440", got)
	}
	wantErr(t, s.SetA4(450.06), ErrInvalidA4)

	c := *appconfig.Get()
	c.Tuner.LockA4 = true
	mustOK(t, appconfig.SetConfig(c))
	wantErr(t, s.SetA4(442), ErrA4Locked)
	if got := s.snapshot().A4; got != 440 {
		t.Fatalf("A4 = %v, want the locked 440", got)
	}
}

// Transpose runs 0..11, and the goal reads errors at the pitch the keys then sound at.
func TestTranspose(t *testing.T) {
	s := freshService(t)
	wantErr(t, s.SetTranspose(12), ErrInvalidTranspose)
	wantErr(t, s.SetTranspose(-1), ErrInvalidTranspose)
	mustOK(t, s.SetTranspose(3))
	if st := s.Settings(); st.Transpose != 3 {
		t.Fatalf("settings transpose = %d, want 3", st.Transpose)
	}
	g := s.adopt()
	if g.a4 != 440 || g.transpose != 3 {
		t.Fatalf("goal = A4 %v / transpose %d, want 440 / 3", g.a4, g.transpose)
	}
	// A reed at the transposed pitch of A4 reads in tune.
	m := measured(69, 0)
	m.Reeds[0].Freq = 440 * math.Exp2(3.0/12)
	if rows := decorate(m, g).ReedErrors; len(rows) != 1 || math.Abs(rows[0].CurrHz) > 1e-9 {
		t.Fatalf("reed rows = %+v, want one at 0 Hz off", rows)
	}
	if got := appconfig.Get().Tuner.Transpose; got != 3 {
		t.Fatalf("config transpose = %d, want 3", got)
	}
}
