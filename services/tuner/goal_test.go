package tuner

import (
	"testing"

	"smegg.me/smeggtuner/core/dsp"
	"smegg.me/smeggtuner/core/tuning"
)

// The goal reaches the frontend on the measurement itself, so the main screen shows it live.
func TestMeasurementCarriesTheGoal(t *testing.T) {
	s, sessions := sessionService(t)
	openSession(t, sessions, 3, 440)
	for reed, cents := range []float64{-8, 0, 8} {
		if err := sessions.SetAnchor(69, reed, cents, "cent"); err != nil {
			t.Fatal(err)
		}
	}

	g := goalOf(s)
	m := measurement(69, -8, 0, 10)
	dto := decorate(m, g)

	if len(dto.ReedErrors) != 3 {
		t.Fatalf("reed errors = %d, want one per reed", len(dto.ReedErrors))
	}
	third := dto.ReedErrors[2]
	if third.Goal != 8 || third.Error < 1.99 || third.Error > 2.01 {
		t.Fatalf("reed 3 = %+v, want a goal of 8 and two cents to come off", third)
	}
	if !dto.ReedErrors[0].InTol {
		t.Fatal("reed 1 sits on its goal and must read as in tune")
	}
	if len(dto.BeatErrors) != 3 {
		t.Fatalf("beat errors = %d, want one per pair", len(dto.BeatErrors))
	}
}

// With no session: every goal is zero, the error is the plain deviation from the scale. A first-class mode, not a fallback.
func TestMeasurementWithNoSessionIsAPureIndicator(t *testing.T) {
	s, _ := sessionService(t)

	dto := decorate(measurement(69, -8, 0, 10), goalOf(s))
	if len(dto.ReedErrors) != 3 {
		t.Fatalf("reed errors = %d, want one per reed even with no goal", len(dto.ReedErrors))
	}
	for _, r := range dto.ReedErrors {
		if r.Goal != 0 || r.Error != r.Curr {
			t.Fatalf("reed %+v, want goal 0 and error == curr with no curve", r)
		}
	}
	// A heartbeat has no reeds and nothing to compare.
	beat := decorate(dsp.Measurement{State: dsp.StateRunning}, goalOf(s))
	if len(beat.ReedErrors) != 0 || len(beat.BeatErrors) != 0 {
		t.Fatal("a heartbeat carries no goal rows")
	}
}

// goalOf is what the emit path reads on every measurement.
func goalOf(s *Service) goal {
	_, _, curve := s.session()
	return goal{curve: curve, a4: s.snapshot().A4}
}

func measurement(note tuning.Note, cents ...float64) dsp.Measurement {
	ref := note.Freq(440)
	m := dsp.Measurement{Note: note, Locked: true, ScalePitch: ref, ReedsSeparated: true, State: dsp.StateRunning}
	for _, c := range cents {
		m.Reeds = append(m.Reeds, dsp.ReedMeasure{Freq: tuning.FreqAtCents(ref, c), DevCents: c})
	}
	return m
}
